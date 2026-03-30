package meta

import (
	"fmt"
	"slices"

	"github.com/openmeterio/openmeter/pkg/models"
	"github.com/qmuntal/stateless"
)

var (
	_             Patch = (*PatchDelete)(nil)
	TriggerDelete       = stateless.Trigger("delete")
)

type PatchDelete struct {
	Policy PatchDeletePolicy
}

func (p PatchDelete) Trigger() stateless.Trigger {
	return TriggerDelete
}

func (p PatchDelete) TriggerParams() any {
	return p.Policy
}

func (p PatchDelete) Validate() error {
	return p.Policy.Validate()
}

type UsageDeletePolicy string

var _ models.Validator = (*UsageDeletePolicy)(nil)

const (
	// UsageDeletePolicyCorrect will refund the usage to the customer by reversing the usage transactions.
	UsageDeletePolicyCorrect UsageDeletePolicy = "correct"
	// UsageDeletePolicyIgnore will ignore the usage and leave it as is without performing any action.
	UsageDeletePolicyIgnore UsageDeletePolicy = "ignore"
)

func (p UsageDeletePolicy) Values() []UsageDeletePolicy {
	return []UsageDeletePolicy{
		UsageDeletePolicyCorrect,
		UsageDeletePolicyIgnore,
	}
}

func (p UsageDeletePolicy) Validate() error {
	if !slices.Contains(p.Values(), p) {
		return fmt.Errorf("invalid credit delete policy: %s", p)
	}

	return nil
}

type PaymentDeletePolicy string

var _ models.Validator = (*PaymentDeletePolicy)(nil)

const (
	// PaymentDeletePolicyRefund will refund the payment to the customer using the app's refund functionality.
	PaymentDeletePolicyRefund PaymentDeletePolicy = "refund"
	// PaymentDeletePolicyGrantCredits will grant credits to the customer to cover the payment amount.
	PaymentDeletePolicyGrantCredits PaymentDeletePolicy = "grant_credits"
	// PaymentDeletePolicyIgnore will ignore the payment and leave it as is without performing any action. (this can be used
	// to settle the payment manually)
	PaymentDeletePolicyIgnore PaymentDeletePolicy = "ignore"
)

func (p PaymentDeletePolicy) Values() []PaymentDeletePolicy {
	return []PaymentDeletePolicy{
		PaymentDeletePolicyRefund,
		PaymentDeletePolicyGrantCredits,
		PaymentDeletePolicyIgnore,
	}
}

func (p PaymentDeletePolicy) Validate() error {
	if !slices.Contains(p.Values(), p) {
		return fmt.Errorf("invalid payment delete policy: %s", p)
	}

	return nil
}

var _ models.Validator = (*PatchDeletePolicy)(nil)

type PatchDeletePolicy struct {
	CreditDeletePolicy         UsageDeletePolicy
	InvoiceAccruedDeletePolicy UsageDeletePolicy
	PaymentDeletePolicy        PaymentDeletePolicy
}

func (p PatchDeletePolicy) Validate() error {
	var errs []error

	if err := p.CreditDeletePolicy.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("credit delete policy: %w", err))
	}

	if err := p.InvoiceAccruedDeletePolicy.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invoice accrued delete policy: %w", err))
	}

	if err := p.PaymentDeletePolicy.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("payment delete policy: %w", err))
	}

	return nil
}

// RefundAsCreditsDeletePolicy is a policy that will refund the usage as credits to the customer. For now this can
// be considered as the default policy for delete patches.
var RefundAsCreditsDeletePolicy PatchDeletePolicy = PatchDeletePolicy{
	CreditDeletePolicy:         UsageDeletePolicyCorrect,
	InvoiceAccruedDeletePolicy: UsageDeletePolicyCorrect,
	PaymentDeletePolicy:        PaymentDeletePolicyGrantCredits,
}
