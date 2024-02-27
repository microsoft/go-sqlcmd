package telemetry

import (
	"os/user"

	"github.com/denisbrodbeck/machineid"
)

func getOrCreateUserId() string {
	user, _ := user.Current()
	id, _ := machineid.ProtectedID(user.Username)
	return id
}
