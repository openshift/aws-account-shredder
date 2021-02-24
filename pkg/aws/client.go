package clientpkg

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

//go:generate mockgen -source=./client.go -destination=../mock/client_generated.go -package=mock

// Client is a wrapper object for actual AWS SDK clients to allow for easier testing.
type Client interface {
	//EC2

	DescribeInstanceStatus(*ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error)
	TerminateInstances(*ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
	DeleteVolume(*ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error)
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeVpcs(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error)
	DeleteVpc(input *ec2.DeleteVpcInput) (*ec2.DeleteVpcOutput, error)
	DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error)
	DeleteSubnet(input *ec2.DeleteSubnetInput) (*ec2.DeleteSubnetOutput, error)
	DescribeInternetGateways(input *ec2.DescribeInternetGatewaysInput) (*ec2.DescribeInternetGatewaysOutput, error)
	DetachInternetGateway(input *ec2.DetachInternetGatewayInput) (*ec2.DetachInternetGatewayOutput, error)
	DescribeNetworkInterfaces(input *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error)
	DetachNetworkInterface(input *ec2.DetachNetworkInterfaceInput) (*ec2.DetachNetworkInterfaceOutput, error)
	DeleteNetworkInterface(input *ec2.DeleteNetworkInterfaceInput) (*ec2.DeleteNetworkInterfaceOutput, error)
	DescribeNatGateways(input *ec2.DescribeNatGatewaysInput) (*ec2.DescribeNatGatewaysOutput, error)
	DeleteNatGateway(input *ec2.DeleteNatGatewayInput) (*ec2.DeleteNatGatewayOutput, error)
	DescribeRouteTables(input *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error)
	DeleteRouteTable(input *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error)
	DescribeNetworkAcls(input *ec2.DescribeNetworkAclsInput) (*ec2.DescribeNetworkAclsOutput, error)
	DeleteNetworkAcl(input *ec2.DeleteNetworkAclInput) (*ec2.DeleteNetworkAclOutput, error)
	DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error)
	DeleteSecurityGroup(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error)
	DeleteInternetGateway(input *ec2.DeleteInternetGatewayInput) (*ec2.DeleteInternetGatewayOutput, error)
	DeleteVpcEndpoints(input *ec2.DeleteVpcEndpointsInput) (*ec2.DeleteVpcEndpointsOutput, error)
	DescribeVpcEndpoints(input *ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error)
	DisassociateRouteTable(input *ec2.DisassociateRouteTableInput) (*ec2.DisassociateRouteTableOutput, error)
	RevokeSecurityGroupIngress(input *ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error)
	DescribeVpnConnections(input *ec2.DescribeVpnConnectionsInput) (*ec2.DescribeVpnConnectionsOutput, error)
	DescribeVpnGateways(input *ec2.DescribeVpnGatewaysInput) (*ec2.DescribeVpnGatewaysOutput, error)
	DeleteVpnConnection(input *ec2.DeleteVpnConnectionInput) (*ec2.DeleteVpnConnectionOutput, error)
	DeleteVpnGateway(input *ec2.DeleteVpnGatewayInput) (*ec2.DeleteVpnGatewayOutput, error)
	DetachVpnGateway(input *ec2.DetachVpnGatewayInput) (*ec2.DetachVpnGatewayOutput, error)
	DescribeSnapshots(input *ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error)
	DeleteSnapshot(input *ec2.DeleteSnapshotInput) (*ec2.DeleteSnapshotOutput, error)
	DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)

	//efs
	DescribeMountTargets(input *efs.DescribeMountTargetsInput) (*efs.DescribeMountTargetsOutput, error)
	DeleteMountTarget(input *efs.DeleteMountTargetInput) (*efs.DeleteMountTargetOutput, error)
	DescribeFileSystems(input *efs.DescribeFileSystemsInput) (*efs.DescribeFileSystemsOutput, error)
	DeleteFileSystem(input *efs.DeleteFileSystemInput) (*efs.DeleteFileSystemOutput, error)

	//ELB
	DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DeleteLoadBalancer(input *elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error)

	//ELBV2
	DescribeLoadBalancers2(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DeleteLoadBalancer2(input *elbv2.DeleteLoadBalancerInput) (*elbv2.DeleteLoadBalancerOutput, error)

	//STS
	AssumeRole(*sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error)
	GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)

	// S3
	ListBuckets(*s3.ListBucketsInput) (*s3.ListBucketsOutput, error)
	DeleteBucket(*s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)
	BatchDeleteBucketObjects(bucketName *string) error

	// Route53
	ListHostedZones(*route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error)
	DeleteHostedZone(*route53.DeleteHostedZoneInput) (*route53.DeleteHostedZoneOutput, error)
	ListResourceRecordSets(*route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(*route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
	GetRegion() string
}

type awsClient struct {
	region        string
	ec2Client     ec2iface.EC2API
	stsClient     stsiface.STSAPI
	s3Client      s3iface.S3API
	route53client route53iface.Route53API
	elbClient     elbiface.ELBAPI
	elbv2Client   elbv2iface.ELBV2API
	efsClient     efsiface.EFSAPI
}

func (c *awsClient) DescribeInstanceStatus(input *ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error) {
	return c.ec2Client.DescribeInstanceStatus(input)
}

func (c *awsClient) TerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	return c.ec2Client.TerminateInstances(input)
}

func (c *awsClient) DeleteVolume(input *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
	return c.ec2Client.DeleteVolume(input)
}

func (c *awsClient) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return c.ec2Client.DescribeInstances(input)
}

func (c *awsClient) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	return c.ec2Client.DescribeVpcs(input)
}

func (c *awsClient) DeleteVpc(input *ec2.DeleteVpcInput) (*ec2.DeleteVpcOutput, error) {
	return c.ec2Client.DeleteVpc(input)
}

func (c *awsClient) DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	return c.ec2Client.DescribeSubnets(input)
}

func (c *awsClient) DeleteSubnet(input *ec2.DeleteSubnetInput) (*ec2.DeleteSubnetOutput, error) {
	return c.ec2Client.DeleteSubnet(input)
}

func (c *awsClient) DescribeInternetGateways(input *ec2.DescribeInternetGatewaysInput) (*ec2.DescribeInternetGatewaysOutput, error) {
	return c.ec2Client.DescribeInternetGateways(input)
}

func (c *awsClient) DetachInternetGateway(input *ec2.DetachInternetGatewayInput) (*ec2.DetachInternetGatewayOutput, error) {
	return c.ec2Client.DetachInternetGateway(input)
}

func (c *awsClient) DescribeNetworkInterfaces(input *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return c.ec2Client.DescribeNetworkInterfaces(input)
}
func (c *awsClient) DetachNetworkInterface(input *ec2.DetachNetworkInterfaceInput) (*ec2.DetachNetworkInterfaceOutput, error) {
	return c.ec2Client.DetachNetworkInterface(input)

}

func (c *awsClient) DeleteNetworkInterface(input *ec2.DeleteNetworkInterfaceInput) (*ec2.DeleteNetworkInterfaceOutput, error) {
	return c.ec2Client.DeleteNetworkInterface(input)
}

func (c *awsClient) DescribeNatGateways(input *ec2.DescribeNatGatewaysInput) (*ec2.DescribeNatGatewaysOutput, error) {
	return c.ec2Client.DescribeNatGateways(input)
}

func (c *awsClient) DeleteNatGateway(input *ec2.DeleteNatGatewayInput) (*ec2.DeleteNatGatewayOutput, error) {
	return c.ec2Client.DeleteNatGateway(input)
}

func (c *awsClient) DescribeRouteTables(input *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
	return c.ec2Client.DescribeRouteTables(input)
}
func (c *awsClient) DeleteRouteTable(input *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error) {
	return c.ec2Client.DeleteRouteTable(input)
}

func (c *awsClient) DescribeNetworkAcls(input *ec2.DescribeNetworkAclsInput) (*ec2.DescribeNetworkAclsOutput, error) {
	return c.ec2Client.DescribeNetworkAcls(input)
}

func (c *awsClient) DeleteNetworkAcl(input *ec2.DeleteNetworkAclInput) (*ec2.DeleteNetworkAclOutput, error) {
	return c.ec2Client.DeleteNetworkAcl(input)
}

func (c *awsClient) DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	return c.ec2Client.DescribeSecurityGroups(input)
}

func (c *awsClient) DeleteSecurityGroup(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
	return c.ec2Client.DeleteSecurityGroup(input)
}

func (c *awsClient) DeleteInternetGateway(input *ec2.DeleteInternetGatewayInput) (*ec2.DeleteInternetGatewayOutput, error) {
	return c.ec2Client.DeleteInternetGateway(input)
}

func (c *awsClient) DescribeVpcEndpoints(input *ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error) {
	return c.ec2Client.DescribeVpcEndpoints(input)
}

func (c *awsClient) DeleteVpcEndpoints(input *ec2.DeleteVpcEndpointsInput) (*ec2.DeleteVpcEndpointsOutput, error) {
	return c.ec2Client.DeleteVpcEndpoints(input)
}

func (c *awsClient) DisassociateRouteTable(input *ec2.DisassociateRouteTableInput) (*ec2.DisassociateRouteTableOutput, error) {
	return c.ec2Client.DisassociateRouteTable(input)
}

func (c *awsClient) RevokeSecurityGroupIngress(input *ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	return c.ec2Client.RevokeSecurityGroupIngress(input)
}

func (c *awsClient) DescribeVpnConnections(input *ec2.DescribeVpnConnectionsInput) (*ec2.DescribeVpnConnectionsOutput, error) {
	return c.ec2Client.DescribeVpnConnections(input)
}

func (c *awsClient) DescribeVpnGateways(input *ec2.DescribeVpnGatewaysInput) (*ec2.DescribeVpnGatewaysOutput, error) {
	return c.ec2Client.DescribeVpnGateways(input)
}

func (c *awsClient) DeleteVpnConnection(input *ec2.DeleteVpnConnectionInput) (*ec2.DeleteVpnConnectionOutput, error) {
	return c.ec2Client.DeleteVpnConnection(input)
}
func (c *awsClient) DeleteVpnGateway(input *ec2.DeleteVpnGatewayInput) (*ec2.DeleteVpnGatewayOutput, error) {
	return c.ec2Client.DeleteVpnGateway(input)
}
func (c *awsClient) DetachVpnGateway(input *ec2.DetachVpnGatewayInput) (*ec2.DetachVpnGatewayOutput, error) {
	return c.ec2Client.DetachVpnGateway(input)
}

func (c *awsClient) DescribeSnapshots(input *ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error) {
	return c.ec2Client.DescribeSnapshots(input)
}

func (c *awsClient) DeleteSnapshot(input *ec2.DeleteSnapshotInput) (*ec2.DeleteSnapshotOutput, error) {
	return c.ec2Client.DeleteSnapshot(input)
}

func (c *awsClient) DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	return c.ec2Client.DescribeVolumes(input)
}

//efs
func (c *awsClient) DescribeMountTargets(input *efs.DescribeMountTargetsInput) (*efs.DescribeMountTargetsOutput, error) {
	return c.efsClient.DescribeMountTargets(input)
}

func (c *awsClient) DeleteMountTarget(input *efs.DeleteMountTargetInput) (*efs.DeleteMountTargetOutput, error) {
	return c.efsClient.DeleteMountTarget(input)
}

func (c *awsClient) DescribeFileSystems(input *efs.DescribeFileSystemsInput) (*efs.DescribeFileSystemsOutput, error) {
	return c.efsClient.DescribeFileSystems(input)
}

func (c *awsClient) DeleteFileSystem(input *efs.DeleteFileSystemInput) (*efs.DeleteFileSystemOutput, error) {
	return c.efsClient.DeleteFileSystem(input)
}

// ELB
func (c *awsClient) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return c.elbClient.DescribeLoadBalancers(input)
}
func (c *awsClient) DeleteLoadBalancer(input *elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error) {
	return c.elbClient.DeleteLoadBalancer(input)
}

//ELB v2

func (c *awsClient) DescribeLoadBalancers2(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
	return c.elbv2Client.DescribeLoadBalancers(input)
}
func (c *awsClient) DeleteLoadBalancer2(input *elbv2.DeleteLoadBalancerInput) (*elbv2.DeleteLoadBalancerOutput, error) {
	return c.elbv2Client.DeleteLoadBalancer(input)
}

func (c *awsClient) AssumeRole(input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	return c.stsClient.AssumeRole(input)
}

func (c *awsClient) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	return c.stsClient.GetCallerIdentity(input)
}

func (c *awsClient) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return c.s3Client.ListBuckets(input)
}

func (c *awsClient) DeleteBucket(input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
	return c.s3Client.DeleteBucket(input)
}

func (c *awsClient) ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return c.s3Client.ListObjectsV2(input)
}

func (c *awsClient) BatchDeleteBucketObjects(bucketName *string) error {
	// Setup BatchDeleteItrerator to iterate through a list of objects
	iter := s3manager.NewDeleteListIterator(c.s3Client, &s3.ListObjectsInput{
		Bucket: bucketName,
	})

	// Traverse iterator deleting each object
	return s3manager.NewBatchDeleteWithClient(c.s3Client).Delete(aws.BackgroundContext(), iter)
}

func (c *awsClient) ListHostedZones(input *route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error) {
	return c.route53client.ListHostedZones(input)
}

func (c *awsClient) DeleteHostedZone(input *route53.DeleteHostedZoneInput) (*route53.DeleteHostedZoneOutput, error) {
	return c.route53client.DeleteHostedZone(input)
}

func (c *awsClient) ListResourceRecordSets(input *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {
	return c.route53client.ListResourceRecordSets(input)
}

func (c *awsClient) ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	return c.route53client.ChangeResourceRecordSets(input)
}

func (c *awsClient) GetRegion() string {
	return c.region
}

// NewClient creates our client wrapper object for the actual AWS clients we use.
func NewClient(awsAccessID, awsAccessSecret, token, region string) (Client, error) {
	awsConfig := &aws.Config{Region: aws.String(region)}
	awsConfig.Credentials = credentials.NewStaticCredentials(
		awsAccessID, awsAccessSecret, token)

	s, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	return &awsClient{
		region:        region,
		ec2Client:     ec2.New(s),
		stsClient:     sts.New(s),
		s3Client:      s3.New(s),
		route53client: route53.New(s),
		elbClient:     elb.New(s),
		elbv2Client:   elbv2.New(s),
		efsClient:     efs.New(s),
	}, nil
}
