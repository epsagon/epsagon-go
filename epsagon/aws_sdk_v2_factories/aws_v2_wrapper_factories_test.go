package epsagonawssdk2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const TEST_ACCOUNT = "test_account"
const TEST_USER_ID = "test_user_id"
const TEST_ARN = "test_arn"

type CallerIdentityMock struct {
	Account *string
	Arn     *string
	UserId  *string
}

var _ = Describe("aws_sdk_v2_factories", func() {
	Describe("sts_factory", func() {
		Context("Happy Flows", func() {
			var (
				request  *aws.Request
				resource *protocol.Resource
			)
			BeforeEach(func() {
				request = &aws.Request{}
				resource = &protocol.Resource{
					Metadata:  map[string]string{},
					Operation: "GetCallerIdentity",
				}
			})
			It("Metadata Only is false, partial data", func() {
				account_data := TEST_ACCOUNT
				request.Data = &CallerIdentityMock{
					Account: &account_data,
				}
				stsDataFactory(request, resource, false, tracer.GlobalTracer)
				Expect(resource.Metadata["Account"]).To(Equal(TEST_ACCOUNT))
			})
			It("Metadata Only is false, full data", func() {
				account_data := TEST_ACCOUNT
				user_id_data := TEST_USER_ID
				arn_data := TEST_ARN
				request.Data = &CallerIdentityMock{
					Account: &account_data,
					Arn:     &arn_data,
					UserId:  &user_id_data,
				}
				stsDataFactory(request, resource, false, tracer.GlobalTracer)
				Expect(len(resource.Metadata)).To(Equal(3))
				Expect(resource.Metadata["Account"]).To(Equal(TEST_ACCOUNT))
				Expect(resource.Metadata["Arn"]).To(Equal(TEST_ARN))
				Expect(resource.Metadata["UserId"]).To(Equal(TEST_USER_ID))
			})
			It("Metadata Only is true", func() {
				account_data := TEST_ACCOUNT
				user_id_data := TEST_USER_ID
				arn_data := TEST_ARN
				request.Data = &CallerIdentityMock{
					Account: &account_data,
					Arn:     &arn_data,
					UserId:  &user_id_data,
				}
				stsDataFactory(request, resource, true, tracer.GlobalTracer)
				Expect(len(resource.Metadata)).To(Equal(0))
			})
		})
	})
})
