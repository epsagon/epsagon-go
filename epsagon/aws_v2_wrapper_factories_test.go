package epsagon

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/epsagon/epsagon-go/epsagon/aws_sdk_v2_factories"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

const TestAccount = "test_account"
const TestUserID = "test_user_id"
const TestArn = "test_arn"

type CallerIdentityMock struct {
	Account *string
	Arn     *string
	UserID  *string
}

func TestEpsagonFactoriesTracer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS SDK V2 Factories")
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
				accountData := TestAccount
				request.Data = &CallerIdentityMock{
					Account: &accountData,
				}
				epsagonawsv2factories.StsDataFactory(request, resource, false, tracer.GlobalTracer)
				Expect(resource.Metadata["Account"]).To(Equal(TestAccount))
			})
			It("Metadata Only is false, full data", func() {
				accountData := TestAccount
				userIDData := TestUserID
				arnData := TestArn
				request.Data = &CallerIdentityMock{
					Account: &accountData,
					Arn:     &arnData,
					UserID:  &userIDData,
				}
				epsagonawsv2factories.StsDataFactory(request, resource, false, tracer.GlobalTracer)
				Expect(len(resource.Metadata)).To(Equal(3))
				Expect(resource.Metadata["Account"]).To(Equal(TestAccount))
				Expect(resource.Metadata["Arn"]).To(Equal(TestArn))
				Expect(resource.Metadata["UserID"]).To(Equal(TestUserID))
			})
			It("Metadata Only is true", func() {
				accountData := TestAccount
				userIDData := TestUserID
				arnData := TestArn
				request.Data = &CallerIdentityMock{
					Account: &accountData,
					Arn:     &arnData,
					UserID:  &userIDData,
				}
				epsagonawsv2factories.StsDataFactory(request, resource, true, tracer.GlobalTracer)
				Expect(len(resource.Metadata)).To(Equal(0))
			})
		})
	})
})
