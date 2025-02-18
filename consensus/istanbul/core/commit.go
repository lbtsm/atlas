// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"errors"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/consensus/istanbul"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
)

// maxValidators represents the maximum number of validators the SNARK circuit supports
// The prover code will then pad any proofs to this maximum to ensure consistent proof structure
// TODO: Eventually make this governable
const maxValidators = uint32(150)

func (c *core) sendCommit() {
	logger := c.newLogger("func", "sendCommit")
	logger.Trace("Sending commit")
	sub := c.current.Subject()
	c.broadcastCommit(sub)
}

func (c *core) generateCommittedSeal(sub *istanbul.Subject) (blscrypto.SerializedSignature, error) {
	fork, cur := new(big.Int).Set(c.backend.ChainConfig().BN256ForkBlock), new(big.Int).Set(sub.View.Sequence)
	seal := PrepareCommittedSeal(sub.Digest, sub.View.Round)
	committedSeal, err := c.backend.SignBLS(seal, []byte{}, false, false, fork, cur)
	if err != nil {
		return blscrypto.SerializedSignature{}, err
	}
	return committedSeal, nil
}

// Generates serialized epoch data for use in the Plumo SNARK circuit.
// Block number and hash may be information for a pending block.
func (c *core) generateEpochValidatorSetData(blockNumber uint64, round uint8, blockHash common.Hash, newValSet istanbul.ValidatorSet) ([]byte, []byte, bool, error) {
	if !istanbul.IsLastBlockOfEpoch(blockNumber, c.config.Epoch) {
		return nil, nil, false, errNotLastBlockInEpoch
	}

	// Serialize the public keys for the validators in the validator set.
	blsPubKeys := []blscrypto.SerializedPublicKey{}
	for _, v := range newValSet.List() {
		blsPubKeys = append(blsPubKeys, v.BLSPublicKey())
	}

	maxNonSigners := uint32(newValSet.Size() - newValSet.MinQuorumSize())

	// Before the Donut fork, use the snark data encoding with epoch entropy.

	// Retrieve the block hash for the last block of the previous epoch.
	parentEpochBlockHash := c.backend.HashForBlock(blockNumber - c.config.Epoch)
	if blockNumber > 0 && parentEpochBlockHash == (common.Hash{}) {
		return nil, nil, false, errors.New("unknown block")
	}

	maxNonSigners = maxValidators - uint32(newValSet.MinQuorumSize())
	message, extraData, err := blscrypto.CryptoType().EncodeEpochSnarkDataCIP22(
		blsPubKeys, maxNonSigners, maxValidators,
		uint16(istanbul.GetEpochNumber(blockNumber, c.config.Epoch)),
		round,
		blscrypto.EpochEntropyFromHash(blockHash),
		blscrypto.EpochEntropyFromHash(parentEpochBlockHash),
	)
	// This is after the Donut hardfork, so signify this uses CIP22.
	return message, extraData, true, err
}

func (c *core) broadcastCommit(sub *istanbul.Subject) {
	logger := c.newLogger("func", "broadcastCommit")

	fork, cur := new(big.Int).Set(c.backend.ChainConfig().BN256ForkBlock), new(big.Int).Set(sub.View.Sequence)
	committedSeal, err := c.generateCommittedSeal(sub)
	if err != nil {
		logger.Error("Failed to commit seal", "err", err)
		return
	}

	currentBlockNumber := c.current.Proposal().Number().Uint64()
	newValSet, err := c.backend.NextBlockValidators(c.current.Proposal())
	if err != nil {
		logger.Error("Failed to get next block's validators", "err", err)
		return
	}
	epochValidatorSetData, epochValidatorSetExtraData, cip22, err := c.generateEpochValidatorSetData(currentBlockNumber, uint8(sub.View.Round.Uint64()), sub.Digest, newValSet)
	if err != nil && err != errNotLastBlockInEpoch {
		logger.Error("Failed to create epoch validator set data", "err", err)
		return
	}
	var epochValidatorSetSeal blscrypto.SerializedSignature
	if err == nil {
		epochValidatorSetSeal, err = c.backend.SignBLS(epochValidatorSetData, epochValidatorSetExtraData, true, cip22, fork, cur)
		if err != nil {
			logger.Error("Failed to sign epoch validator set seal", "err", err)
			return
		}
	}
	istMsg := istanbul.NewCommitMessage(&istanbul.CommittedSubject{
		Subject:               sub,
		CommittedSeal:         committedSeal[:],
		EpochValidatorSetSeal: epochValidatorSetSeal[:],
	}, c.address)
	c.broadcast(istMsg, false)
}

func (c *core) handleCommit(msg *istanbul.Message) error {
	flag := c.assembleMsgFlag(msg)
	log.Info("receipt commit message ------------------", "sender", msg.Address, "cur_seq", c.current.Sequence(), "msg_seq", msg.Commit().Subject.View.Sequence)
	_, ok := c.forwardedMap[flag]
	if ok { // is forward it handler
		c.logger.Info("handleCommit this msg is handled", "flag", flag)
		return nil
	}

	if !c.config.Validator {
		c.logger.Info("not validator, only forward", "address", c.address)
		c.forwardCommit(msg)
		return nil
	}
	defer c.handleCommitTimer.UpdateSince(time.Now())
	commit := msg.Commit()
	err := c.checkMessage(istanbul.MsgCommit, commit.Subject.View)
	if err == errOldMessage {
		// Discard messages from previous views, unless they are commits from the previous sequence,
		// with the same round as what we wound up finalizing, as we would be able to include those
		// to create the ParentAggregatedSeal for our next proposal.
		lastSubject, err := c.backend.LastSubject()
		if err != nil {
			return err
		} else if commit.Subject.View.Cmp(lastSubject.View) != 0 {
			return errOldMessage
		} else if lastSubject.View.Sequence.Cmp(common.Big0) == 0 {
			// Don't handle commits for the genesis block, will cause underflows
			return errOldMessage
		}
		return c.handleCheckedCommitForPreviousSequence(msg, commit)
	} else if err != nil {
		return err
	}

	return c.handleCheckedCommitForCurrentSequence(msg, commit)
}

func (c *core) handleCheckedCommitForPreviousSequence(msg *istanbul.Message, commit *istanbul.CommittedSubject) error {
	logger := c.newLogger("func", "handleCheckedCommitForPreviousSequence", "tag", "handleMsg", "msg_view", commit.Subject.View)
	headBlock := c.backend.GetCurrentHeadBlock()
	// Retrieve the validator set for the previous proposal (which should
	// match the one broadcast)
	parentValset := c.backend.ParentBlockValidators(headBlock)
	_, validator := parentValset.GetByAddress(msg.Address)
	if validator == nil {
		return errInvalidValidatorAddress
	}
	fork, cur := new(big.Int).Set(c.backend.ChainConfig().BN256ForkBlock), new(big.Int).Set(headBlock.Number())
	if err := c.verifyCommittedSeal(commit, validator, fork, cur); err != nil {
		log.Info("handleCheckedCommitForPreviousSequence -------------- ", "cur_seq", c.current.Sequence(), "msg_seq", msg.Commit().Subject.View.Sequence, "sender", msg.Address)
		return errInvalidCommittedSeal
	}
	if headBlock.Number().Uint64() > 0 {
		if err := c.verifyEpochValidatorSetSeal(commit, headBlock.Number().Uint64(), c.current.ValidatorSet(), validator); err != nil {
			return errInvalidEpochValidatorSetSeal
		}
	}

	// Ensure that the commit's digest (ie the received proposal's hash) matches the head block's hash
	if headBlock.Number().Uint64() > 0 && commit.Subject.Digest != headBlock.Hash() {
		logger.Debug("Received a commit message for the previous sequence with an unexpected hash", "expected", headBlock.Hash().String(), "received", commit.Subject.Digest.String())
		return errInconsistentSubject
	}

	// Add the ParentCommit to current round state
	if err := c.current.AddParentCommit(msg); err != nil {
		logger.Error("Failed to record parent seal", "m", msg, "err", err)
		return err
	}
	return nil
}

func (c *core) handleCheckedCommitForCurrentSequence(msg *istanbul.Message, commit *istanbul.CommittedSubject) error {
	logger := c.newLogger("func", "handleCheckedCommitForCurrentSequence", "tag", "handleMsg")
	validator := c.current.GetValidatorByAddress(msg.Address)
	if validator == nil {
		return errInvalidValidatorAddress
	}

	fork, cur := new(big.Int).Set(c.backend.ChainConfig().BN256ForkBlock), new(big.Int).Set(c.current.Proposal().Number())
	if err := c.verifyCommittedSeal(commit, validator, fork, cur); err != nil {
		log.Info("handleCheckedCommitForCurrentSequence ++++++++++++++++", "cur_seq", c.current.Sequence(), "msg_seq", msg.Commit().Subject.View.Sequence, "sender", msg.Address)
		return errInvalidCommittedSeal
	}

	newValSet, err := c.backend.NextBlockValidators(c.current.Proposal())
	if err != nil {
		return err
	}

	if err := c.verifyEpochValidatorSetSeal(commit, c.current.Proposal().Number().Uint64(), newValSet, validator); err != nil {
		return errInvalidEpochValidatorSetSeal
	}

	// ensure that the commit is in the current proposal
	if err := c.verifyCommit(commit); err != nil {
		return err
	}

	// Add the COMMIT message to current round state
	if err := c.current.AddCommit(msg); err != nil {
		logger.Error("Failed to record commit message", "m", msg, "err", err)
		return err
	}
	numberOfCommits := c.current.Commits().Size()
	minQuorumSize := c.current.ValidatorSet().MinQuorumSize()
	logger.Trace("Accepted commit for current sequence", "Number of commits", numberOfCommits, "minQuorumSize", minQuorumSize)

	// Commit the proposal once we have enough COMMIT messages and we are not in the Committed state.
	//
	// If we already have a proposal, we may have chance to speed up the consensus process
	// by committing the proposal without PREPARE messages.
	// TODO(joshua): Remove state comparisons (or change the cmp function)
	if numberOfCommits >= minQuorumSize && c.current.State().Cmp(StateCommitted) < 0 {
		logger.Trace("Got a quorum of commits", "tag", "stateTransition", "commits", numberOfCommits, "quorum", minQuorumSize)
		err := c.commit()
		if err != nil {
			logger.Error("Failed to commit()", "err", err)
			return err
		}
		c.forwardedMap = make(map[string]struct{}) // make it empty

	} else if c.current.GetPrepareOrCommitSize() >= minQuorumSize && c.current.State().Cmp(StatePrepared) < 0 {
		err := c.current.TransitionToPrepared(minQuorumSize)
		if err != nil {
			logger.Error("Failed to create and set prepared certificate", "err", err)
			return err
		}
		// Process Backlog Messages
		c.backlog.updateState(c.current.View(), c.current.State())

		logger.Info("Got quorum prepares or commits", "tag", "stateTransition", "commits", c.current.Commits,
			"prepares", c.current.Prepares, "c.current.GetPrepareOrCommitSize()", c.current.GetPrepareOrCommitSize(), "c.current.State()", c.current.State())
		c.sendCommit()
	}

	if msg.Commit().Subject.View.Sequence.Cmp(c.current.Sequence()) < 0 {
		logger.Info("Not Need forward commit", "cur_seq", c.current.Sequence(), "msg_seq", msg.Commit().Subject.View.Sequence)
		return nil
	}
	logger.Info("forward commit", "cur_seq", c.current.Sequence(), "msg_seq", msg.Commit().Subject.View.Sequence, "sender", msg.Address)
	c.forwardCommit(msg)
	return nil

}

// verifyCommit verifies if the received COMMIT message is equivalent to our subject
func (c *core) verifyCommit(commit *istanbul.CommittedSubject) error {
	logger := c.newLogger("func", "verifyCommit")

	sub := c.current.Subject()
	if !reflect.DeepEqual(commit.Subject, sub) {
		logger.Warn("Inconsistent subjects between commit and proposal", "expected", sub, "got", commit)
		return errInconsistentSubject
	}

	return nil
}

// verifyCommittedSeal verifies the commit seal in the received COMMIT message
func (c *core) verifyCommittedSeal(comSub *istanbul.CommittedSubject, src istanbul.Validator, fork, cur *big.Int) error {
	seal := PrepareCommittedSeal(comSub.Subject.Digest, comSub.Subject.View.Round)
	return blscrypto.CryptoType().VerifySignature(src.BLSPublicKey(), seal, []byte{}, comSub.CommittedSeal,
		false, false, fork, cur)
}

// verifyEpochValidatorSetSeal verifies the epoch validator set seal in the received COMMIT message
func (c *core) verifyEpochValidatorSetSeal(comSub *istanbul.CommittedSubject, blockNumber uint64,
	newValSet istanbul.ValidatorSet, src istanbul.Validator) error {
	if blockNumber == 0 {
		return nil
	}
	epochData, epochExtraData, cip22, err := c.generateEpochValidatorSetData(blockNumber, uint8(comSub.Subject.View.Round.Uint64()), comSub.Subject.Digest, newValSet)
	if err != nil {
		if err == errNotLastBlockInEpoch {
			return nil
		}
		return err
	}
	fork, cur := new(big.Int).Set(c.backend.ChainConfig().BN256ForkBlock), big.NewInt(int64(blockNumber))
	return blscrypto.CryptoType().VerifySignature(src.BLSPublicKey(), epochData, epochExtraData,
		comSub.EpochValidatorSetSeal, true, cip22, fork, cur)
}

func (c *core) forwardCommit(msg *istanbul.Message) {
	istMsg := istanbul.NewCommitMessage(msg.Commit(), msg.Address)
	c.broadcast(istMsg, true)
}
