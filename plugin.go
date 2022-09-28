package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

type Plugin struct {
	Key         []byte
	ContainerId string
}

func (p *Plugin) Exec() error {
	key, err := iamkey.ReadFromJSONBytes(p.Key)
	if err != nil {
		return fmt.Errorf("cannot parse service account key: %w", err)
	}

	creds, err := ycsdk.ServiceAccountKey(key)
	if err != nil {
		return fmt.Errorf("failed to onbtain SDK credentiasl: %w", err)
	}

	ctx := context.Background()

	Info("Initializing SDK")
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{Credentials: creds})
	if err != nil {
		return fmt.Errorf("failed to initialize SDK: %w", err)
	}

	f := logrus.Fields{"container": p.ContainerId}

	Info("Getting current revision", logrus.Fields{"container": p.ContainerId})
	revs, err := sdk.Serverless().Containers().Container().ListRevisions(ctx, &containers.ListContainersRevisionsRequest{
		Id: &containers.ListContainersRevisionsRequest_ContainerId{ContainerId: p.ContainerId},
	})
	if err != nil {
		return WithFields(fmt.Errorf("failed to fetch current revision: %w", err), f)
	}
	if len(revs.Revisions) == 0 {
		return WithFields(fmt.Errorf("did not find existing revisions"), f)
	}
	rev := revs.Revisions[0]

	if strings.Contains(rev.Image.ImageUrl, "@") {
		f["image"] = rev.Image.ImageUrl
		return WithFields(fmt.Errorf("deployed image url must not be pinned to revision"), f)
	}

	Info("Deploying new revision", logrus.Fields{"container": p.ContainerId})
	if _, err = sdk.Serverless().Containers().Container().DeployRevision(ctx,
		&containers.DeployContainerRevisionRequest{
			ContainerId:      p.ContainerId,
			Description:      rev.Description,
			Resources:        rev.Resources,
			ExecutionTimeout: rev.ExecutionTimeout,
			ServiceAccountId: rev.ServiceAccountId,
			ImageSpec: &containers.ImageSpec{
				ImageUrl:    rev.Image.ImageUrl,
				Command:     rev.Image.Command,
				Args:        rev.Image.Args,
				Environment: rev.Image.Environment,
				WorkingDir:  rev.Image.WorkingDir,
			},
			Concurrency:     rev.Concurrency,
			Secrets:         rev.Secrets,
			Connectivity:    rev.Connectivity,
			ProvisionPolicy: rev.ProvisionPolicy,
		}); err != nil {
		return WithFields(fmt.Errorf("failed to deploy new revision: %w", err), f)
	}

	Info("Success", f)
	return nil
}
