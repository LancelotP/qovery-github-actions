package qovery

import (
	"fmt"
	"strings"
	"time"

	"github-action/pkg"
)

func DeployDatabase(qoveryAPIClient pkg.QoveryAPIClient, databaseId string, qoveryEnvironmentId string) error {
	timeout := time.Second * 1800 // 30 minutes

	// Checking deployment is not QUEUED or DEPLOYING already
	// if so, wait for it to be ready
	stateIsOk := false
	var status *pkg.EnvironmentStatus
	for start := time.Now(); time.Since(start) < timeout; {
		status, err := qoveryAPIClient.GetEnvironmentStatus(qoveryEnvironmentId)
		if err != nil {
			return fmt.Errorf("error while trying to get environment status: %s", err)
		}

		// Statuses ok to start a deployment
		if status.State == pkg.EnvStatusDeploymentError ||
			status.State == pkg.EnvStatusStopError ||
			status.State == pkg.EnvStatusRunning ||
			status.State == pkg.EnvStatusRunningError ||
			status.State == pkg.EnvStatusCancelError ||
			status.State == pkg.EnvStatusCancelled ||
			status.State == pkg.EnvStatusUnknown {
			stateIsOk = true
			break
		}

		fmt.Printf("Environment cannot accept deploy yet, state: %s\n", status.State)

		time.Sleep(10 * time.Second)
	}

	// Environment state is not valid even after timeout, cannot deploy the database
	if !stateIsOk {
		return fmt.Errorf("error: database cannot be deployed, environment status is : %s", status.State)
	}

	// Launching deployment
	err := qoveryAPIClient.DeployDatabase(pkg.Database{ID: databaseId})
	if err != nil {
		return fmt.Errorf("error while trying to deploy database: %s", err)
	}

	// Waiting for deployment to be OK or ERRORED with a timeout
	for start := time.Now(); time.Since(start) < timeout; {
		status, err := qoveryAPIClient.GetEnvironmentStatus(qoveryEnvironmentId)
		if err != nil {
			return fmt.Errorf("error while trying to get environment status: %s", err)
		}

		fmt.Printf("Deployment ongoing: status %s\n", status.State)

		if status.State == pkg.EnvStatusRunning {
			break
		} else if strings.HasSuffix(string(status.State), "ERROR") {
			return fmt.Errorf("error: database has not been deployed, environment status is : %s", status.State)
		}

		time.Sleep(10 * time.Second)
	}

	return nil
}
