// Copyright © 2021 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package multinode

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

// SubmitSyncCommitteeContributions submits sync committee contributions.
func (s *Service) SubmitSyncCommitteeContributions(ctx context.Context, contributionAndProofs []*altair.SignedContributionAndProof) error {
	if len(contributionAndProofs) == 0 {
		return errors.New("no contribution and proofs supplied")
	}

	sem := semaphore.NewWeighted(s.processConcurrency)
	var wg sync.WaitGroup
	for name, submitter := range s.syncCommitteeContributionsSubmitters {
		wg.Add(1)
		go func(ctx context.Context,
			sem *semaphore.Weighted,
			wg *sync.WaitGroup,
			name string,
			submitter eth2client.SyncCommitteeContributionsSubmitter,
		) {
			defer wg.Done()
			log := log.With().Str("beacon_node_address", name).Uint64("slot", uint64(contributionAndProofs[0].Message.Contribution.Slot)).Logger()
			if err := sem.Acquire(ctx, 1); err != nil {
				log.Error().Err(err).Msg("Failed to acquire semaphore")
				return
			}
			defer sem.Release(1)

			_, address := s.serviceInfo(ctx, submitter)
			started := time.Now()
			err := submitter.SubmitSyncCommitteeContributions(ctx, contributionAndProofs)
			s.clientMonitor.ClientOperation(address, "submit contribution and proofs", err == nil, time.Since(started))
			if err != nil {
				log.Warn().Err(err).Msg("Failed to submit contribution and proofs")
			} else {
				data, err := json.Marshal(contributionAndProofs)
				if err != nil {
					log.Error().Err(err).Msg("Failed to marshal JSON")
				} else {
					log.Trace().Str("data", string(data)).Msg("Submitted contribution and proofs")
				}
			}
		}(ctx, sem, &wg, name, submitter)
	}
	wg.Wait()

	return nil
}
