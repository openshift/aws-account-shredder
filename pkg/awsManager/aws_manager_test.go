package awsManager

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-logr/logr"

	"github.com/golang/mock/gomock"
	"github.com/openshift/aws-account-shredder/pkg/localMetrics"
	"github.com/openshift/aws-account-shredder/pkg/mock"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type mockSuite struct {
	mockCtrl      *gomock.Controller
	mockAWSClient *mock.MockClient
	Logger        logr.Logger
}

// setupDefaultMocks is an easy way to setup all of the default mocks
func setupDefaultMocks(t *testing.T) *mockSuite {
	mocks := &mockSuite{
		mockCtrl: gomock.NewController(t),
		Logger:   logf.Log.WithName("shredder_mock_logger"),
	}

	mocks.mockAWSClient = mock.NewMockClient(mocks.mockCtrl)
	return mocks
}

func init() {
	localMetrics.Initialize("", "")
}

func TestDeleteS3Buckets(t *testing.T) {
	testCases := []struct {
		title         string
		setupAWSMock  func(r *mock.MockClientMockRecorder)
		listOfBuckets []*string
		errorExpected bool
	}{
		{
			title: "test 1 - No buckets passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {

			},
			listOfBuckets: nil,
			errorExpected: false,
		}, {
			title: "test 2 - Invalid Buckets passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.BatchDeleteBucketObjects(gomock.Any()).Return(nil).AnyTimes()
				r.DeleteBucket(gomock.Any()).Return(&s3.DeleteBucketOutput{}, errors.New("ERROR")).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			listOfBuckets: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected: true,
		}, {
			title: "test 3 - valid Buckets passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.BatchDeleteBucketObjects(gomock.Any()).Return(nil).AnyTimes()
				r.DeleteBucket(gomock.Any()).Return(&s3.DeleteBucketOutput{}, nil).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			listOfBuckets: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteS3Buckets(mocks.mockAWSClient, tc.listOfBuckets, mocks.Logger)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}

func TestDeleteEc2Instance(t *testing.T) {

	testCases := []struct {
		title           string
		setupAWSMock    func(r *mock.MockClientMockRecorder)
		listOfInstances []*string
		errorExpected   bool
	}{
		{
			title: "test 1 - No Instances passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {

			},
			listOfInstances: nil,
			errorExpected:   false,
		}, {
			title: "test 2 - Invalid Instances passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.TerminateInstances(gomock.Any()).Return(&ec2.TerminateInstancesOutput{}, errors.New("Error")).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			listOfInstances: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:   true,
		}, {
			title: "test 3 - valid Instances passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.TerminateInstances(gomock.Any()).Return(&ec2.TerminateInstancesOutput{}, nil).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			listOfInstances: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteEc2Instance(mocks.mockAWSClient, tc.listOfInstances, mocks.Logger)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}

func TestCleanUpAwsRoute53(t *testing.T) {
	testCases := []struct {
		title         string
		setupAWSMock  func(r *mock.MockClientMockRecorder)
		errorExpected bool
	}{
		{
			title: "test 1 - unable to list hosted zone in that region",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.ListHostedZones(gomock.Any()).Return(&route53.ListHostedZonesOutput{}, errors.New("unable to lst hosted zones")).AnyTimes()
			},
			errorExpected: true,
		}, {
			title: "test 2 - Hosted zones are available in that region , able to list hosted zones",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.ListHostedZones(gomock.Any()).Return(&route53.ListHostedZonesOutput{IsTruncated: aws.Bool(false)}, nil).AnyTimes()
				r.ListResourceRecordSets(gomock.Any()).Return(&route53.ListResourceRecordSetsOutput{}, nil).AnyTimes()
				r.ChangeResourceRecordSets(gomock.Any()).Return(&route53.ChangeResourceRecordSetsOutput{}, nil).AnyTimes()
				r.DeleteHostedZone(gomock.Any()).Return(&route53.DeleteHostedZoneOutput{}, nil).AnyTimes()
			},
			errorExpected: true,
		},
		// NOTE :
		// other test cases can be generated by changing the above 4 function calls
		// please change the return() parameter accordingly
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := CleanUpAwsRoute53(mocks.mockAWSClient, mocks.Logger)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}

func TestDeleteVpcInstacnes(t *testing.T) {
	testCases := []struct {
		title           string
		setupAWSMock    func(r *mock.MockClientMockRecorder)
		listOfInstances []*string
		errorExpected   bool
	}{
		{
			title: "test 1 - All the resource clear up in the VPC",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DescribeVpcEndpoints(gomock.Any()).Return(&ec2.DescribeVpcEndpointsOutput{}, nil).AnyTimes()
				r.DeleteVpcEndpoints(gomock.Any()).Return(&ec2.DeleteVpcEndpointsOutput{}, nil).AnyTimes()
				r.DescribeLoadBalancers(gomock.Any()).Return(&elb.DescribeLoadBalancersOutput{}, nil).AnyTimes()
				r.DeleteLoadBalancer(gomock.Any()).Return(&elb.DeleteLoadBalancerOutput{}, nil).AnyTimes()
				r.DescribeLoadBalancers2(gomock.Any()).Return(&elbv2.DescribeLoadBalancersOutput{}, nil).AnyTimes()
				r.DeleteLoadBalancer2(gomock.Any()).Return(&elbv2.DeleteLoadBalancerOutput{}, nil).AnyTimes()
				r.DescribeNatGateways(gomock.Any()).Return(&ec2.DescribeNatGatewaysOutput{}, nil).AnyTimes()
				r.DeleteNatGateway(gomock.Any()).Return(&ec2.DeleteNatGatewayOutput{}, nil).AnyTimes()
				r.DescribeNetworkInterfaces(gomock.Any()).Return(&ec2.DescribeNetworkInterfacesOutput{}, nil).AnyTimes()
				r.DetachNetworkInterface(gomock.Any()).Return(&ec2.DetachNetworkInterfaceOutput{}, nil).AnyTimes()
				r.DeleteNetworkInterface(gomock.Any()).Return(&ec2.DeleteNetworkInterfaceOutput{}, nil).AnyTimes()
				r.DescribeInternetGateways(gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{}, nil).AnyTimes()
				r.DeleteInternetGateway(gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil).AnyTimes()
				r.DescribeVpnGateways(gomock.Any()).Return(&ec2.DescribeVpnGatewaysOutput{}, nil).AnyTimes()
				r.DescribeNetworkAcls(gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{}, nil).AnyTimes()
				r.DeleteNetworkAcl(gomock.Any()).Return(&ec2.DeleteNetworkAclOutput{}, nil).AnyTimes()
				r.DescribeRouteTables(gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{}, nil).AnyTimes()
				r.DisassociateRouteTable(gomock.Any()).Return(&ec2.DisassociateRouteTableOutput{}, nil).AnyTimes()
				r.DeleteRouteTable(gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).AnyTimes()
				r.DescribeSubnets(gomock.Any()).Return(&ec2.DescribeSubnetsOutput{}, nil).AnyTimes()
				r.DeleteSubnet(gomock.Any()).Return(&ec2.DeleteSubnetOutput{}, nil).AnyTimes()
				r.DescribeSecurityGroups(gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{}, nil).AnyTimes()
				r.RevokeSecurityGroupIngress(gomock.Any()).Return(&ec2.RevokeSecurityGroupIngressOutput{}, nil).AnyTimes()
				r.DeleteSecurityGroup(gomock.Any()).Return(&ec2.DeleteSecurityGroupOutput{}, nil).AnyTimes()
				r.DeleteVpc(gomock.Any()).Return(&ec2.DeleteVpcOutput{}, nil).AnyTimes()
			},
			listOfInstances: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:   false,
		}, {
			title: "test 2 - All the resources dont clear up in the VPC",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DescribeVpcEndpoints(gomock.Any()).Return(&ec2.DescribeVpcEndpointsOutput{}, nil).AnyTimes()
				r.DeleteVpcEndpoints(gomock.Any()).Return(&ec2.DeleteVpcEndpointsOutput{}, nil).AnyTimes()
				r.DescribeLoadBalancers(gomock.Any()).Return(&elb.DescribeLoadBalancersOutput{}, nil).AnyTimes()
				r.DeleteLoadBalancer(gomock.Any()).Return(&elb.DeleteLoadBalancerOutput{}, nil).AnyTimes()
				r.DescribeLoadBalancers2(gomock.Any()).Return(&elbv2.DescribeLoadBalancersOutput{}, nil).AnyTimes()
				r.DeleteLoadBalancer2(gomock.Any()).Return(&elbv2.DeleteLoadBalancerOutput{}, errors.New("Error")).AnyTimes()
				r.DescribeNatGateways(gomock.Any()).Return(&ec2.DescribeNatGatewaysOutput{}, nil).AnyTimes()
				r.DeleteNatGateway(gomock.Any()).Return(&ec2.DeleteNatGatewayOutput{}, nil).AnyTimes()
				r.DescribeNetworkInterfaces(gomock.Any()).Return(&ec2.DescribeNetworkInterfacesOutput{}, nil).AnyTimes()
				r.DetachNetworkInterface(gomock.Any()).Return(&ec2.DetachNetworkInterfaceOutput{}, nil).AnyTimes()
				r.DeleteNetworkInterface(gomock.Any()).Return(&ec2.DeleteNetworkInterfaceOutput{}, nil).AnyTimes()
				r.DescribeInternetGateways(gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{}, nil).AnyTimes()
				r.DeleteInternetGateway(gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil).AnyTimes()
				r.DescribeVpnGateways(gomock.Any()).Return(&ec2.DescribeVpnGatewaysOutput{}, nil).AnyTimes()
				r.DescribeNetworkAcls(gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{}, nil).AnyTimes()
				r.DeleteNetworkAcl(gomock.Any()).Return(&ec2.DeleteNetworkAclOutput{}, nil).AnyTimes()
				r.DescribeRouteTables(gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{}, nil).AnyTimes()
				r.DisassociateRouteTable(gomock.Any()).Return(&ec2.DisassociateRouteTableOutput{}, nil).AnyTimes()
				r.DeleteRouteTable(gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).AnyTimes()
				r.DescribeSubnets(gomock.Any()).Return(&ec2.DescribeSubnetsOutput{}, nil).AnyTimes()
				r.DeleteSubnet(gomock.Any()).Return(&ec2.DeleteSubnetOutput{}, nil).AnyTimes()
				r.DescribeSecurityGroups(gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{}, nil).AnyTimes()
				r.RevokeSecurityGroupIngress(gomock.Any()).Return(&ec2.RevokeSecurityGroupIngressOutput{}, nil).AnyTimes()
				r.DeleteSecurityGroup(gomock.Any()).Return(&ec2.DeleteSecurityGroupOutput{}, nil).AnyTimes()
				r.DeleteVpc(gomock.Any()).Return(&ec2.DeleteVpcOutput{}, nil).AnyTimes()
			},
			errorExpected:   true,
			listOfInstances: []*string{aws.String("abcd"), aws.String("abcd")},
		},
		// NOTE :
		// other test cases can be generated by changing the above 4 function calls
		// please change the return() parameter accordingly
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteVpcInstances(mocks.mockAWSClient, tc.listOfInstances, mocks.Logger)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}

func TestDeleteEbsSnapshot(t *testing.T) {
	testCases := []struct {
		title              string
		setupAWSMock       func(r *mock.MockClientMockRecorder)
		listOfEbsSnapshots []*string
		errorExpected      bool
	}{
		{
			title: "test 1 - No Instances passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {

			},
			listOfEbsSnapshots: nil,
			errorExpected:      false,
		}, {
			title: "test 2 - Invalid EBS snapshots passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DeleteSnapshot(gomock.Any()).Return(&ec2.DeleteSnapshotOutput{}, errors.New("ERROR")).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			listOfEbsSnapshots: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:      true,
		}, {
			title: "test 3 - valid EBS instance passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DeleteSnapshot(gomock.Any()).Return(&ec2.DeleteSnapshotOutput{}, nil).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			listOfEbsSnapshots: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteEbsSnapshots(mocks.mockAWSClient, tc.listOfEbsSnapshots, mocks.Logger)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}

func TestDeleteEbsVolumes(t *testing.T) {
	testCases := []struct {
		title            string
		setupAWSMock     func(r *mock.MockClientMockRecorder)
		listOfEbsVolumes []*string
		errorExpected    bool
	}{
		{
			title: "test 1 - No EBS volume ID passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {

			},
			listOfEbsVolumes: nil,
			errorExpected:    false,
		}, {
			title: "test 2 - Invalid EBS Volume ID's passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DeleteVolume(gomock.Any()).Return(&ec2.DeleteVolumeOutput{}, errors.New("ERROR")).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			listOfEbsVolumes: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:    true,
		}, {
			title: "test 3 - valid EBS  value ID's instance passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DeleteVolume(gomock.Any()).Return(&ec2.DeleteVolumeOutput{}, nil).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			listOfEbsVolumes: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteEbsVolumes(mocks.mockAWSClient, tc.listOfEbsVolumes, mocks.Logger)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}

func TestDeleteEFS(t *testing.T) {
	testCases := []struct {
		title                 string
		setupAWSMock          func(r *mock.MockClientMockRecorder)
		fileSystemToBeDeleted []*string
		errorExpected         bool
	}{
		{
			title: "test 1 - No EFS ID passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {

			},
			fileSystemToBeDeleted: nil,
			errorExpected:         false,
		}, {
			title: "test 2 - Invalid EFS ID passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DeleteFileSystem(gomock.Any()).Return(&efs.DeleteFileSystemOutput{}, errors.New("Error")).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			fileSystemToBeDeleted: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:         true,
		}, {
			title: "test 3 - valid EFS ID passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DeleteFileSystem(gomock.Any()).Return(&efs.DeleteFileSystemOutput{}, nil).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			fileSystemToBeDeleted: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteEFS(mocks.mockAWSClient, tc.fileSystemToBeDeleted, mocks.Logger)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}

func TestDeleteEFSMountTarget(t *testing.T) {
	testCases := []struct {
		title                  string
		setupAWSMock           func(r *mock.MockClientMockRecorder)
		mountTargetToBeDeleted []*string
		errorExpected          bool
	}{
		{
			title: "test 1 - No EFS mount target ID passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {

			},
			mountTargetToBeDeleted: nil,
			errorExpected:          false,
		}, {
			title: "test 2 - Invalid EFS mount target ID passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DeleteMountTarget(gomock.Any()).Return(&efs.DeleteMountTargetOutput{}, errors.New("ERROR")).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			mountTargetToBeDeleted: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:          true,
		}, {
			title: "test 3 - valid EFS mount target ID passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DeleteMountTarget(gomock.Any()).Return(&efs.DeleteMountTargetOutput{}, nil).AnyTimes()
				r.GetRegion().Return("Region1").AnyTimes()
			},
			mountTargetToBeDeleted: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:          false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteEFSMountTarget(mocks.mockAWSClient, tc.mountTargetToBeDeleted, mocks.Logger)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}

func TestCleanEIPAddresses(t *testing.T) {

	failedToRetrieveEIPErr := errors.New("FailedToRetrieveEIP")
	failedToReleaseAddressErr := errors.New("FailedToReleaseAddress")

	testCases := []struct {
		title         string
		setupAWSMock  func(r *mock.MockClientMockRecorder)
		eipAddresses  []*string
		errorExpected error
	}{
		{
			title: "No EIP Address",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DescribeAddresses(gomock.Any()).Return(&ec2.DescribeAddressesOutput{}, nil).AnyTimes()
			},
			errorExpected: nil,
		},
		{
			title: "EIP Address No Error",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DescribeAddresses(gomock.Any()).Return(&ec2.DescribeAddressesOutput{Addresses: []*ec2.Address{
					{
						AllocationId: aws.String("test-id-one"),
					},
					{
						AllocationId: aws.String("test-id-two"),
					},
				}}, nil).AnyTimes()
				r.ReleaseAddress(gomock.Any()).Return(&ec2.ReleaseAddressOutput{}, nil).AnyTimes()
			},
			errorExpected: nil,
		},
		{
			title: "Get EIP Addresses Error",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DescribeAddresses(gomock.Any()).Return(&ec2.DescribeAddressesOutput{}, failedToRetrieveEIPErr).AnyTimes()
			},
			errorExpected: failedToRetrieveEIPErr,
		},
		{
			title: "Release Address Error",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.DescribeAddresses(gomock.Any()).Return(&ec2.DescribeAddressesOutput{Addresses: []*ec2.Address{
					{
						AllocationId: aws.String("test-id-one"),
					},
					{
						AllocationId: aws.String("test-id-two"),
					},
				}}, nil).AnyTimes()
				r.ReleaseAddress(gomock.Any()).Return(&ec2.ReleaseAddressOutput{}, failedToReleaseAddressErr).AnyTimes()
			},
			errorExpected: failedToReleaseAddressErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())
			mockExecution := CleanEIPAddresses(mocks.mockAWSClient, mocks.Logger)
			if mockExecution != tc.errorExpected {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}
