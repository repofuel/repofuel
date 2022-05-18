package invoke

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

//go:generate stringer -type=Action -linecomment  -trimprefix Action
//go:generate jsonenums -type=Action
type Action uint8 // todo: consider to rename it to event if it more suitable

const (
	ActionRepositoryAdded         Action = iota + 1 // Repository add
	ActionRepositoryPush                            // Repository push
	ActionRepositoryRecovering                      // Repository recover
	ActionRepositoryAdminTrigger                    // Repository admin trigger
	ActionRepositoryRefreshing                      // Repository refresh
	ActionPullRequestAdded                          // Pull request add
	ActionPullRequestUpdate                         // Pull request update
	ActionPullRequestRecovering                     // Pull request recover
	ActionPullRequestAdminTrigger                   // Pull request admin trigger
	ActionPullRequestRefreshing                     // Pull request refresh

	ActionPushCheck
	ActionPullRequestCheck

	ActionMonitorRepository // Monitor repository
)

var _ActionEnumToActionValue = make(map[string]Action, len(_ActionValueToName))
var _ActionValueToQuotedActionEnumTo = make(map[Action]string, len(_ActionValueToName))

func init() {
	for k := range _ActionValueToName {
		name := strings.ReplaceAll(strings.ToUpper(k.String()), " ", "_")
		_ActionEnumToActionValue[name] = k
		_ActionValueToQuotedActionEnumTo[k] = strconv.Quote(name)
	}

}

func (t *Action) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	tag, ok := _ActionEnumToActionValue[s]
	if !ok {
		return fmt.Errorf("invalid Action %q", s)
	}
	*t = tag
	return nil
}

func (t Action) MarshalGQL(w io.Writer) {
	v, ok := _ActionValueToQuotedActionEnumTo[t]
	if !ok {
		fmt.Fprint(w, strconv.Quote(t.String()))
		return
	}

	fmt.Fprint(w, v)
}
