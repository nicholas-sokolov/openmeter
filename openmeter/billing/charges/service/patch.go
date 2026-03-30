package service

import (
	"context"
	"fmt"

	"github.com/openmeterio/openmeter/openmeter/billing/charges"
	"github.com/openmeterio/openmeter/openmeter/customer"
	"github.com/openmeterio/openmeter/pkg/framework/transaction"
	"github.com/samber/lo"
)

func (s *service) ApplyPatches(ctx context.Context, input charges.ApplyPatchesInput) (charges.Charges, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	return transaction.Run(ctx, s.adapter, func(ctx context.Context) (charges.Charges, error) {
		// TODO: Is this a good response type?
		out := make(charges.Charges, 0, len(input.PatchesByChargeID))

		patchedCharges, err := s.applyPatches(ctx, input.CustomerID, input.PatchesByChargeID)
		if err != nil {
			return nil, err
		}
		out = append(out, patchedCharges...)

		if len(input.Creates) > 0 {
			// Charge creation is the last step as patches might delete a charge whose UniqueReferenceID is used in the creation.
			createdCharges, err := s.Create(ctx, charges.CreateInput{
				Namespace: input.CustomerID.Namespace,
				Intents:   input.Creates,
			})
			if err != nil {
				return nil, err
			}

			out = append(out, createdCharges...)
		}

		return out, nil
	})
}

func (s *service) applyPatches(ctx context.Context, customerID customer.CustomerID, patchesByChargeID map[string]charges.Patch) (charges.Charges, error) {
	chargesItems, err := s.adapter.GetByIDs(ctx, charges.GetByIDsInput{
		Namespace: customerID.Namespace,
		IDs:       lo.Keys(patchesByChargeID),
	})
	if err != nil {
		return nil, err
	}

	// Let's validate the charges items
	for _, charge := range chargesItems {
		if charge.CustomerID != customerID.ID {
			return nil, fmt.Errorf("charge %s is not owned by customer %s", charge.ID.ID, customerID.ID)
		}

		if charge.ID.Namespace != customerID.Namespace {
			return nil, fmt.Errorf("charge %s is not in namespace %s, expected %s", charge.ID.ID, customerID.Namespace, charge.ID.Namespace)
		}
	}

	invocableChargesByID, err := s.newInvocableCharges(chargesItems)
	if err != nil {
		return nil, err
	}

	out := make(charges.Charges, 0, len(patchesByChargeID))

	for chargeID, patch := range patchesByChargeID {
		invocableCharge, ok := invocableChargesByID[chargeID]
		if !ok {
			return nil, fmt.Errorf("charge %s not found", chargeID)
		}

		patchedCharge, err := invocableCharge.TriggerPatch(ctx, patch)
		if err != nil {
			return nil, err
		}

		if patchedCharge == nil {
			// TODO: is this an error?!
			continue
		}

		out = append(out, *patchedCharge)
	}

	return out, nil
}
