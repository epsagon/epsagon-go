package epsagon

import (

	awsFactories "github.com/epsagon/epsagon-go/epsagon/aws_sdk_v2_factories"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const testAccount = "test_account"
const testUserID = "test_user_id"
const testARN = "test_arn"

type CallerIdentityMock struct {
	Account *string
	Arn     *string
	UserId  *string
}

var _ = Describe("aws_sdk_v2_factories", func() {
	Describe("sts_factory", func() {
		Context("Happy Flows", func() {
			var (
				r  *awsFactories.AWSCall
				resource *protocol.Resource
			)
			BeforeEach(func() {
				r = &awsFactories.AWSCall{}
				resource = &protocol.Resource{
					Metadata:  map[string]string{},
					Operation: "GetCallerIdentity",
				}
			})
			It("Metadata Only is false, partial data", func() {
				accountData := testAccount
				r.Output = &CallerIdentityMock{
					Account: &accountData,
				}
				awsFactories.StsEventDataFactory(r, resource, false, tracer.GlobalTracer)
				Expect(resource.Metadata["Account"]).To(Equal(testAccount))
			})
			It("Metadata Only is false, full data", func() {
				accountData := testAccount
				userIDData := testUserID
				arnData := testARN
				r.Output = &CallerIdentityMock{
					Account: &accountData,
					Arn:     &arnData,
					UserId:  &userIDData,
				}
				awsFactories.StsEventDataFactory(r, resource, false, tracer.GlobalTracer)
				Expect(len(resource.Metadata)).To(Equal(3))
				Expect(resource.Metadata["Account"]).To(Equal(testAccount))
				Expect(resource.Metadata["Arn"]).To(Equal(testARN))
				Expect(resource.Metadata["UserId"]).To(Equal(testUserID))
			})
			It("Metadata Only is true", func() {
				accountData := testAccount
				userIDData := testUserID
				arnData := testARN
				r.Output = &CallerIdentityMock{
					Account: &accountData,
					Arn:     &arnData,
					UserId:  &userIDData,
				}
				awsFactories.StsEventDataFactory(r, resource, true, tracer.GlobalTracer)
				Expect(len(resource.Metadata)).To(Equal(0))
			})
		})
	})
})
