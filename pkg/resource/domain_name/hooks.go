// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package domain_name

import (
	"context"
	"fmt"

	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/apigateway"

	svcapitags "github.com/aws-controllers-k8s/apigateway-controller/pkg/tags"
)

// getDomainNameARN returns the ARN for a given domain name
func (rm *resourceManager) getDomainNameARN(domainName string) string {
	// API Gateway domain name ARN format:
	// arn:aws:apigateway:region::/domainnames/domain-name
	return fmt.Sprintf(
		"arn:%s:apigateway:%s::/domainnames/%s",
		rm.awsAccountID,
		rm.awsRegion,
		domainName,
	)
}

// getTags returns the tags for a given domain name
func (rm *resourceManager) getTags(
	ctx context.Context,
	domainName string,
) map[string]*string {
	// Get the ARN of the domain name
	resourceARN := rm.getDomainNameARN(domainName)

	// Call the GetTags API
	input := &svcsdk.GetTagsInput{
		ResourceArn: &resourceARN,
	}
	resp, err := rm.sdkapi.GetTags(ctx, input)
	rm.metrics.RecordAPICall("GET", "GetTags", err)
	if err != nil {
		return nil
	}

	// Convert map[string]string to map[string]*string
	tags := make(map[string]*string, len(resp.Tags))
	for k, v := range resp.Tags {
		value := v
		tags[k] = &value
	}

	return tags
}

// syncTags synchronizes the tags for a given domain name
func (rm *resourceManager) syncTags(
	ctx context.Context,
	latest *resource,
	desired *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncTags")
	defer func() { exit(err) }()

	// Get the ARN of the domain name
	resourceARN := rm.getDomainNameARN(*desired.ko.Spec.DomainName)

	// Get the existing tags
	existingTags := map[string]*string{}
	if latest != nil && latest.ko.Spec.Tags != nil {
		existingTags = latest.ko.Spec.Tags
	}

	// Get the desired tags
	desiredTags := map[string]*string{}
	if desired.ko.Spec.Tags != nil {
		desiredTags = desired.ko.Spec.Tags
	}

	return svcapitags.SyncTags(
		ctx,
		rm.sdkapi,
		rm.metrics,
		resourceARN,
		desiredTags,
		existingTags,
	)
}
