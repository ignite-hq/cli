package network

import (
	"context"
	"fmt"

	launchtypes "github.com/tendermint/spn/x/launch/types"
	"github.com/tendermint/starport/starport/pkg/events"
	"github.com/tendermint/starport/starport/services/network/networkchain"
)

// LaunchParams fetches the chain launch module params from SPN
func (n Network) LaunchParams(ctx context.Context) (launchtypes.Params, error) {
	res, err := launchtypes.NewQueryClient(n.cosmos.Context).Params(ctx, &launchtypes.QueryParamsRequest{})
	if err != nil {
		return launchtypes.Params{}, err
	}
	return res.GetParams(), nil
}

// TriggerLaunch launches a chain as a coordinator
func (n Network) TriggerLaunch(ctx context.Context, launchID, remainingTime uint64) error {
	n.ev.Send(events.New(events.StatusOngoing, fmt.Sprintf("Launching chain %d", launchID)))

	address := n.account.Address(networkchain.SPN)
	params, err := n.LaunchParams(ctx)
	if err != nil {
		return err
	}

	if remainingTime < params.MinLaunchTime {
		return fmt.Errorf("remaining time %d lower than minimum %d",
			remainingTime,
			params.MaxLaunchTime)
	} else if remainingTime > params.MaxLaunchTime {
		return fmt.Errorf("remaining time %d greater than maximum %d",
			remainingTime,
			params.MaxLaunchTime)
	}

	msg := launchtypes.NewMsgTriggerLaunch(address, launchID, remainingTime)
	n.ev.Send(events.New(events.StatusOngoing, "Broadcasting launch transaction"))
	res, err := n.cosmos.BroadcastTx(n.account.Name, msg)
	if err != nil {
		return err
	}

	var launchRes launchtypes.MsgTriggerLaunchResponse
	if err := res.Decode(&launchRes); err != nil {
		return err
	}

	n.ev.Send(events.New(events.StatusDone,
		fmt.Sprintf("Chain %d will be launched in %d seconds", launchID, remainingTime),
	))
	return nil
}
