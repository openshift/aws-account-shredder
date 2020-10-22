package awsManager

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/go-logr/logr"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"github.com/openshift/aws-account-shredder/pkg/localMetrics"
)

// ErrVpcNotDelete indicates there was an error in the process of deleting a VPCs
var ErrVpcNotDelete = errors.New("VpcNotDelete")

// ListVPCforDeletion returns a list of VPCs suitable for deletion
func ListVPCforDeletion(client clientpkg.Client) ([]*string, error) {

	var vpcToBeDeleted []*string
	var token *string
	for {
		vpcList, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{NextToken: token})
		if err != nil {
			return nil, err
		}

		for _, vpcs := range vpcList.Vpcs {
			if *vpcs.IsDefault == false {
				vpcToBeDeleted = append(vpcToBeDeleted, vpcs.VpcId)
			}
		}

		if vpcList.NextToken != nil {
			token = vpcList.NextToken
		} else {
			break
		}
	}
	return vpcToBeDeleted, nil
}

// CleanVpcInstances lists and removes listed vcp instances
func CleanVpcInstances(client clientpkg.Client, logger logr.Logger) error {
	vpcToBeDeleted, err := ListVPCforDeletion(client)
	if err != nil {
		logger.Error(err, "Failed to list VPCs")
		return err
	}

	err = DeleteVpcInstances(client, vpcToBeDeleted, logger)
	if err != nil {
		logger.Error(err, "Failed to delete VPCs")
		return err
	}

	// need to clear VPN connection now , VPC gateway has been detached already by now (function call in -> deleteVpcInstances()) , if not this will throw an error .
	err = DeleteVpnConnections(client, logger)
	if err != nil {
		logger.Error(err, "Failed to delete VPN connections")
		return err
	}

	logger.Info("VPCs deleted successfully")
	return nil
}

// DeleteVpcInstances deletes all VPCs given
func DeleteVpcInstances(client clientpkg.Client, vpcToBeDeleted []*string, logger logr.Logger) error {

	errFlag := false
	for _, vpcID := range vpcToBeDeleted {

		//need to clean out the dependencies
		// EC2 + S3 are already cleaned out by now
		// delete vpc endpoints
		err := DeleteVpcEndpoint(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete VPC endpoints")
			errFlag = true
		}
		// clear out all ELB
		err = DeleteELB(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete ELBs")
			errFlag = true
		}
		// clear out all network load balancer
		err = DeleteNetworkLoadBalancer(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete Network Load Balancers")
			errFlag = true
		}
		// delete NAT gateway
		err = DeleteNatgateway(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete Nat Gateway")
			errFlag = true
		}
		// detach and delete network interface
		err = DetachAndDeleteNetworkInterface(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to detach and delete Network Interface")
			errFlag = true
		}
		//detach any internet gateway , and delete it
		err = DeleteGateway(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete Gateway")
			errFlag = true
		}
		//detach VPN gateway
		err = DetachVpnGateway(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to detach VPN Gateway")
			errFlag = true
		}
		//cleaning network ACL
		err = DeleteNetworkAcl(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete Network ACL")
			errFlag = true
		}
		//now cleaning route tables
		err = DeleteRouteTables(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete Route Tables")
			errFlag = true
		}
		// now cleaning subnets associated with that vpc id
		err = DeleteSubnetsForVPC(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete Subnet for VPCs")
			errFlag = true
		}
		//now cleaning security groups
		err = DeleteSecurityGroups(client, vpcID, logger)
		if err != nil {
			logger.Error(err, "Failed to delete security groups")
			errFlag = true
		}

		output, err := client.DeleteVpc(&ec2.DeleteVpcInput{VpcId: vpcID})
		if err != nil {
			logger.Error(err, "Failed to delete VPC", output)
			errFlag = true
		}
	}

	if errFlag == true {
		return ErrVpcNotDelete
	}
	logger.Info("VPC's deleted successfully")
	return nil
}

func DeleteELB(client clientpkg.Client, vpcID *string, logger logr.Logger) error {
	var marker *string

	for {
		elbList, err := client.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{Marker: marker})
		if err != nil {
			return err
		}

		for _, elasticLoadBalancer := range elbList.LoadBalancerDescriptions {
			if *elasticLoadBalancer.VPCId == *vpcID {
				_, err := client.DeleteLoadBalancer(&elb.DeleteLoadBalancerInput{LoadBalancerName: elasticLoadBalancer.LoadBalancerName})
				if err != nil {
					logger.Error(err, "Failed to delete ELB", *elasticLoadBalancer.LoadBalancerName)
					localMetrics.ResourceFail(localMetrics.ElasticLoadBalancer, client.GetRegion())
					continue
				}
				localMetrics.ResourceSuccess(localMetrics.ElasticLoadBalancer, client.GetRegion())
			}
		}

		if elbList.NextMarker != nil {
			marker = elbList.NextMarker
		} else {
			break
		}

	}
	return nil
}

func DeleteNatgateway(client clientpkg.Client, vpcID *string, logger logr.Logger) error {
	var token *string

	for {
		natGatewayList, err := client.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{NextToken: token})
		if err != nil {
			return err
		}

		for _, natGateway := range natGatewayList.NatGateways {
			if *natGateway.VpcId == *vpcID {
				_, err := client.DeleteNatGateway(&ec2.DeleteNatGatewayInput{NatGatewayId: natGateway.NatGatewayId})
				if err != nil {
					logger.Error(err, "Failed to delete NAT Gateway", *natGateway.NatGatewayId)
					localMetrics.ResourceFail(localMetrics.NatGateway, client.GetRegion())
					continue
				}
				localMetrics.ResourceSuccess(localMetrics.NatGateway, client.GetRegion())
			}
		}

		if natGatewayList.NextToken != nil {
			token = natGatewayList.NextToken
		} else {
			break
		}
	}
	return nil
}

func DeleteNetworkLoadBalancer(client clientpkg.Client, vpcID *string, logger logr.Logger) error {
	var marker *string

	for {
		networkLoadBalancerList, err := client.DescribeLoadBalancers2(&elbv2.DescribeLoadBalancersInput{Marker: marker})
		if err != nil {
			return err
		}

		for _, networkLoadBalancer := range networkLoadBalancerList.LoadBalancers {
			if *networkLoadBalancer.VpcId == *vpcID {
				_, err := client.DeleteLoadBalancer2(&elbv2.DeleteLoadBalancerInput{LoadBalancerArn: networkLoadBalancer.LoadBalancerArn})
				if err != nil {
					logger.Error(err, "Failed to delete Network Load Balancer", *networkLoadBalancer.LoadBalancerName)
					localMetrics.ResourceFail(localMetrics.NetworkLoadBalancer, client.GetRegion())
				} else {
					localMetrics.ResourceSuccess(localMetrics.NetworkLoadBalancer, client.GetRegion())
				}
			}
		}
		if networkLoadBalancerList.NextMarker != nil {
			marker = networkLoadBalancerList.NextMarker
		} else {
			break
		}
	}
	return nil
}

func DetachAndDeleteNetworkInterface(client clientpkg.Client, vpcID *string, logger logr.Logger) error {
	var token *string

	for {
		networkInterfaceList, err := client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{NextToken: token})
		if err != nil {
			logger.Error(err, "Failed to describe Network Interface")
			return err
		}

		for _, networkInterface := range networkInterfaceList.NetworkInterfaces {
			if networkInterface == nil || networkInterface.VpcId == nil || vpcID == nil {
				continue
			}
			if *networkInterface.VpcId == *vpcID {
				_, err := client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{AttachmentId: networkInterface.NetworkInterfaceId})
				if err != nil {
					logger.Error(err, "Failure to detach interface", *networkInterface.NetworkInterfaceId)
				}

				// delete interface
				_, err = client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{NetworkInterfaceId: networkInterface.NetworkInterfaceId})
				if err != nil {
					logger.Error(err, "Failed to delete Network Interface", *networkInterface.NetworkInterfaceId)
					localMetrics.ResourceFail(localMetrics.NetworkInterface, client.GetRegion())
				} else {
					localMetrics.ResourceSuccess(localMetrics.NetworkInterface, client.GetRegion())
				}
			}
		}
		if networkInterfaceList.NextToken != nil {
			token = networkInterfaceList.NextToken
		} else {
			break
		}
	}
	return nil
}

func DeleteGateway(client clientpkg.Client, vpcID *string, logger logr.Logger) error {
	var token *string

	for {
		internetGatewayList, err := client.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{NextToken: token})
		if err != nil {
			logger.Error(err, "Failed to list Internet Gateways")
			return err
		}

		for _, gateway := range internetGatewayList.InternetGateways {
			for _, attachments := range gateway.Attachments {
				if *attachments.VpcId == *vpcID {
					_, err := client.DetachInternetGateway(&ec2.DetachInternetGatewayInput{InternetGatewayId: gateway.InternetGatewayId, VpcId: attachments.VpcId})
					if err != nil {
						logger.Error(err, "Failed to detach Internet Gateway", *gateway.InternetGatewayId)
					}
					// delete internet gateway
					_, err = client.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{InternetGatewayId: gateway.InternetGatewayId})
					if err != nil {
						logger.Error(err, "Failed to delete Internet Gateway", *gateway.InternetGatewayId)
						localMetrics.ResourceFail(localMetrics.InternetGateway, client.GetRegion())
						continue
					}
					localMetrics.ResourceSuccess(localMetrics.InternetGateway, client.GetRegion())
				}
			}
		}

		if internetGatewayList.NextToken != nil {
			token = internetGatewayList.NextToken
		} else {
			break
		}
	}
	return nil
}

func DeleteSubnetsForVPC(client clientpkg.Client, vpcId *string, logger logr.Logger) error {
	var token *string

	for {
		subnetList, err := client.DescribeSubnets(&ec2.DescribeSubnetsInput{NextToken: token})
		if err != nil {
			logger.Error(err, "Failed to retrieve Subnet list")
			return err
		}

		for _, subnet := range subnetList.Subnets {
			if *subnet.VpcId == *vpcId {
				_, err := client.DeleteSubnet(&ec2.DeleteSubnetInput{SubnetId: subnet.SubnetId})
				if err != nil {
					logger.Error(err, "Failed to delete subnet-id", *subnet.SubnetId)
					localMetrics.ResourceFail(localMetrics.Subnet, client.GetRegion())
					continue
				}
				localMetrics.ResourceSuccess(localMetrics.Subnet, client.GetRegion())
			}
		}

		if subnetList.NextToken != nil {
			token = subnetList.NextToken
		} else {
			break
		}
	}
	return nil
}

func DeleteRouteTables(client clientpkg.Client, vpcId *string, logger logr.Logger) error {
	var token *string

	for {
		routeTableList, err := client.DescribeRouteTables(&ec2.DescribeRouteTablesInput{NextToken: token})
		if err != nil {
			fmt.Println(err, "Failed to retrieve Route Table")
			return err
		}

		for _, routeTable := range routeTableList.RouteTables {
			for _, association := range routeTable.Associations {
				if *routeTable.VpcId == *vpcId && association.RouteTableAssociationId != nil {
					//disassociate route table
					_, err = client.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{AssociationId: association.RouteTableAssociationId})
					if err != nil {
						logger.Error(err, "Failed to disassociate route-table", *routeTable.RouteTableId)
					}
				}
			}
		}

		for _, routeTable := range routeTableList.RouteTables {
			if *routeTable.VpcId == *vpcId {

				_, err = client.DeleteRouteTable(&ec2.DeleteRouteTableInput{RouteTableId: routeTable.RouteTableId})
				if err != nil {
					logger.Error(err, "Failed to delete route-table", *routeTable.RouteTableId)
					localMetrics.ResourceFail(localMetrics.RouteTable, client.GetRegion())
					continue
				}
				localMetrics.ResourceSuccess(localMetrics.RouteTable, client.GetRegion())
			}
		}

		if routeTableList.NextToken != nil {
			token = routeTableList.NextToken
		} else {
			break
		}
	}

	return nil
}

func DeleteNetworkAcl(client clientpkg.Client, vpcId *string, logger logr.Logger) error {

	var token *string

	for {
		aclList, err := client.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{NextToken: token})
		if err != nil {
			logger.Error(err, "Failed to retrieve ACL")
			return err
		}

		for _, acl := range aclList.NetworkAcls {
			if *acl.VpcId == *vpcId {
				_, err := client.DeleteNetworkAcl(&ec2.DeleteNetworkAclInput{NetworkAclId: acl.NetworkAclId})
				if err != nil {
					logger.Error(err, "Failed to delete ACL", *acl.NetworkAclId)
					localMetrics.ResourceFail(localMetrics.NetworkACL, client.GetRegion())
					continue
				}
				localMetrics.ResourceSuccess(localMetrics.NetworkACL, client.GetRegion())
			}
		}

		if aclList.NextToken != nil {
			token = aclList.NextToken
		} else {
			break
		}
	}

	return nil
}

func DeleteSecurityGroups(client clientpkg.Client, vpcId *string, logger logr.Logger) error {

	var token *string

	for {
		securityGroupList, err := client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{NextToken: token})
		if err != nil {
			logger.Error(err, "Failed to retrieve Security Group list")
			return err
		}

		for _, securityGroup := range securityGroupList.SecurityGroups {
			if *securityGroup.VpcId == *vpcId {
				_, err = client.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{IpPermissions: securityGroup.IpPermissions, GroupId: securityGroup.GroupId})
				if err != nil {
					logger.Error(err, "Failed to delete all permissions")
				}

			}
		}

		for _, securityGroup := range securityGroupList.SecurityGroups {
			if *securityGroup.VpcId == *vpcId {
				_, err = client.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{GroupId: securityGroup.GroupId})
				if err != nil {
					logger.Error(err, "Failed to delete Security Group", *securityGroup.GroupId)
					localMetrics.ResourceFail(localMetrics.SecurityGroup, client.GetRegion())
					continue
				}
				localMetrics.ResourceSuccess(localMetrics.SecurityGroup, client.GetRegion())
			}
		}
		if securityGroupList.NextToken != nil {
			token = securityGroupList.NextToken
		} else {
			break
		}
	}

	return nil
}

func DeleteVpcEndpoint(client clientpkg.Client, vpcId *string, logger logr.Logger) error {

	var vpcEndpointToBeDeleted []*string
	var token *string
	for {
		vpcEndpointList, err := client.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{NextToken: token})
		if err != nil {
			logger.Error(err, "Failed to retrieve VPC endpoint list")
			return err
		}

		for _, vpcEndpoint := range vpcEndpointList.VpcEndpoints {
			if *vpcEndpoint.VpcId == *vpcId {
				vpcEndpointToBeDeleted = append(vpcEndpointToBeDeleted, vpcEndpoint.VpcEndpointId)
			}
		}

		if vpcEndpointList.NextToken != nil {
			token = vpcEndpointList.NextToken
		} else {
			break
		}
	}

	if vpcEndpointToBeDeleted == nil {
		return nil
	}

	vpcNotDeleted, err := client.DeleteVpcEndpoints(&ec2.DeleteVpcEndpointsInput{VpcEndpointIds: vpcEndpointToBeDeleted})
	if err != nil {
		logger.Error(err, "Failed to delete VPC", vpcNotDeleted.String())
		localMetrics.ResourceFail(localMetrics.VPC, client.GetRegion())
	} else {
		logger.Info("ALL VPCs have been deleted successfully")
		localMetrics.ResourceSuccess(localMetrics.VPC, client.GetRegion())
	}
	return nil
}

func DeleteVpnConnections(client clientpkg.Client, logger logr.Logger) error {

	// does not require pagination
	vpnConnectionList, err := client.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{})
	if err != nil {
		logger.Error(err, "Failed to retrieve VPN connection list")
		return err
	}

	for _, vpnConnection := range vpnConnectionList.VpnConnections {
		_, err = client.DeleteVpnConnection(&ec2.DeleteVpnConnectionInput{VpnConnectionId: vpnConnection.VpnConnectionId})
		if err != nil {
			logger.Error(err, "Failed to delete VPN connection", *vpnConnection.VpnConnectionId)
			localMetrics.ResourceFail(localMetrics.VpnConnection, client.GetRegion())
			continue
		}
		localMetrics.ResourceSuccess(localMetrics.VpnConnection, client.GetRegion())
	}

	return nil

}

func DetachVpnGateway(client clientpkg.Client, vpcId *string, logger logr.Logger) error {

	// does not require pagination
	vpnGatewayList, err := client.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{})
	if err != nil {
		logger.Error(err, "Failed to retrieve VPN Gateway list")
		return err
	}

	for _, vpnGateway := range vpnGatewayList.VpnGateways {
		_, err = client.DetachVpnGateway(&ec2.DetachVpnGatewayInput{VpcId: vpcId, VpnGatewayId: vpnGateway.VpnGatewayId})
		if err != nil {
			logger.Error(err, "Failed to detach VPN gateway", *vpnGateway.VpnGatewayId)
			localMetrics.ResourceFail(localMetrics.VpnGateway, client.GetRegion())
			continue
		}
		localMetrics.ResourceFail(localMetrics.VpnGateway, client.GetRegion())
	}

	return nil

}
