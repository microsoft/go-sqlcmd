// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package container

import (
	"context"
	"net/http"

	"github.com/distribution/reference"
	"github.com/docker/distribution/registry/client"
)

// ListTags lists all tags for a container image located at a given
// path in the container registry. It takes the path to the image and the
// URL of the registry as input and returns a slice of strings containing
// the tags.
func ListTags(path string, baseURL string) []string {
	ctx := context.Background()
	repo, err := reference.WithName(path)
	checkErr(err)
	repository, err := client.NewRepository(
		repo,
		baseURL,
		http.DefaultTransport,
	)
	checkErr(err)
	tagService := repository.Tags(ctx)
	tags, err := tagService.All(ctx)
	checkErr(err)

	return tags
}
