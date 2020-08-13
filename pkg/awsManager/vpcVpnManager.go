package awsManager

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
)

var ErrVpcNotDelete = errors.New("VpcNotDelete")

func ListVPCforDeletion(client clientpkg.Client) ([]*string, error) {

	var vpcToBeDeleted []*string
	var token *string
	for {
		vpcList, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{NextToken: token})
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Println("Error: Cant list VPC ")
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

func CleanVpcInstances(client clientpkg.Client) error {
	vpcToBeDeleted, err := ListVPCforDeletion(client)
	if err != nil {
		return err
	}

	err = DeleteVpcInstances(client, vpcToBeDeleted)

	if err != nil {
		return err
	}
	// need to clear VPN connection now , VPC gateway has been detached already by now (function call in -> deleteVpcInstances()) , if not this will throw an error .
	err = DeleteVpnConnections(client)
	if err != nil {
		return err
	}

	return nil
}

func DeleteVpcInstances(client clientpkg.Client, vpcToBeDeleted []*string) error {

	errFlag := false
	for _, vpcId := range vpcToBeDeleted {

		//need to clean out the dependencies
		// EC2 + S3 are already cleaned out by now
		// delete vpc endpoints
		err := DeleteVpcEndpoint(client, vpcId)
		if err != nil {
			errFlag = true
		}
		// clear out all ELB
		err = DeleteELB(client, vpcId)
		if err != nil {
			errFlag = true
		}
		// clear out all network load balancer
		err = DeleteNetworkLoadBalancer(client, vpcId)
		if err != nil {
			errFlag = true
		}
		// delete NAT gateway
		err = DeleteNatgateway(client, vpcId)
		if err != nil {
			errFlag = true
		}
		// detach and delete network interface
		err = DetachAndDeleteNetworkInterface(client, vpcId)
		if err != nil {
			errFlag = true
		}
		//detach any internet gateway , and delete it
		err = DeleteGateway(client, vpcId)
		if err != nil {
			errFlag = true
		}
		//detach VPN gateway
		err = DetachVpnGateway(client, vpcId)
		if err != nil {
			errFlag = true
		}
		//cleaning network ACL
		err = DeleteNetworkAcl(client, vpcId)
		if err != nil {
			errFlag = true
		}
		//now cleaning route tables
		err = DeleteRouteTables(client, vpcId)
		if err != nil {
			errFlag = true
		}
		// now cleaning subnets associated with that vpc id
		err = DeleteSubnetsForVPC(client, vpcId)
		if err != nil {
			errFlag = true
		}
		//now cleaning security groups
		err = DeleteSecurityGroups(client, vpcId)
		if err != nil {
			errFlag = true
		}

		_, err = client.DeleteVpc(&ec2.DeleteVpcInput{VpcId: vpcId})
		if err != nil {
			errFlag = true
			fmt.Println("ERROR deleting VPC :", *vpcId)
			fmt.Println(err)
		}
	}

	if errFlag == true {
		return ErrVpcNotDelete
	}
	fmt.Println("VPC's deleted successfully for this region  ")
	return nil

}

func DeleteELB(client clientpkg.Client, vpcID *string) error {

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
					fmt.Println("ERROR: cant delete ELB", *elasticLoadBalancer.LoadBalancerName)
					fmt.Println(err)

				}
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

func DeleteNatgateway(client clientpkg.Client, vpcID *string) error {

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
					fmt.Println("ERROR: cant delete NAT GATEWAY ", *natGateway.NatGatewayId)
					fmt.Println(err)

				}

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

func DeleteNetworkLoadBalancer(client clientpkg.Client, vpcID *string) error {

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
					fmt.Println("ERROR: cant delete network load balancer", *networkLoadBalancer.LoadBalancerName)
					fmt.Println(err)

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

func DetachAndDeleteNetworkInterface(client clientpkg.Client, vpcID *string) error {
	var token *string

	for {
		networkInterfaceList, err := client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{NextToken: token})
		if err != nil {
			fmt.Println("ERROR: could not list network interfaces")
			fmt.Println(err)

			return err
		}

		for _, networkInterface := range networkInterfaceList.NetworkInterfaces {
			if networkInterface == nil || networkInterface.VpcId == nil || vpcID == nil {
				continue
			}
			if *networkInterface.VpcId == *vpcID {
				_, err := client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{AttachmentId: networkInterface.NetworkInterfaceId})
				if err != nil {
					fmt.Println("ERROR: cant detach interface", *networkInterface.NetworkInterfaceId)
					fmt.Println(err)
				}

				// delete interface
				_, err = client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{NetworkInterfaceId: networkInterface.NetworkInterfaceId})
				if err != nil {
					fmt.Println("ERROR: cant delete interface", *networkInterface.NetworkInterfaceId)
					fmt.Println(err)

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

func DeleteGateway(client clientpkg.Client, vpcID *string) error {

	var token *string

	for {
		internetGatewayList, err := client.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{NextToken: token})
		if err != nil {
			fmt.Println("ERROR : cant list internet gateways")
			fmt.Println(err)
			return err
		}

		for _, gateway := range internetGatewayList.InternetGateways {
			for _, attachments := range gateway.Attachments {
				if *attachments.VpcId == *vpcID {
					_, err := client.DetachInternetGateway(&ec2.DetachInternetGatewayInput{InternetGatewayId: gateway.InternetGatewayId, VpcId: attachments.VpcId})
					if err != nil {
						fmt.Println("ERROR: cant detach internet gateway", *gateway.InternetGatewayId)
						fmt.Println(err)

					}
					// delete internet gateway
					_, err = client.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{InternetGatewayId: gateway.InternetGatewayId})
					if err != nil {
						fmt.Println("ERROR: cant delete internet gateway", *gateway.InternetGatewayId)
						fmt.Println(err)

					}

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

func DeleteSubnetsForVPC(client clientpkg.Client, vpcId *string) error {

	var token *string

	for {
		subnetList, err := client.DescribeSubnets(&ec2.DescribeSubnetsInput{NextToken: token})
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR: Unable to retreive subnet list ")
			return err
		}

		for _, subnet := range subnetList.Subnets {
			if *subnet.VpcId == *vpcId {
				_, err := client.DeleteSubnet(&ec2.DeleteSubnetInput{SubnetId: subnet.SubnetId})
				if err != nil {
					fmt.Println("ERROR : cant delete the subnet-id ", *subnet.SubnetId)
					fmt.Println(err)

				}
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

func DeleteRouteTables(client clientpkg.Client, vpcId *string) error {

	var token *string

	for {
		routeTableList, err := client.DescribeRouteTables(&ec2.DescribeRouteTablesInput{NextToken: token})
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR: Unable to retreive route table ")
			return err
		}

		for _, routeTable := range routeTableList.RouteTables {
			for _, association := range routeTable.Associations {
				if *routeTable.VpcId == *vpcId && association.RouteTableAssociationId != nil {
					//disassociate route table
					_, err = client.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{AssociationId: association.RouteTableAssociationId})
					if err != nil {
						fmt.Println("ERROR : cant disassociate route-table ", *routeTable.RouteTableId)
						fmt.Println(err)

					}
				}
			}
		}

		for _, routeTable := range routeTableList.RouteTables {
			if *routeTable.VpcId == *vpcId {

				_, err = client.DeleteRouteTable(&ec2.DeleteRouteTableInput{RouteTableId: routeTable.RouteTableId})
				if err != nil {
					fmt.Println("ERROR : cant delete route-table ", *routeTable.RouteTableId)
					fmt.Println(err)
				}
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

func DeleteNetworkAcl(client clientpkg.Client, vpcId *string) error {

	var token *string

	for {
		aclList, err := client.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{NextToken: token})
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR: Unable to retreive ACL ")
			return err
		}

		for _, acl := range aclList.NetworkAcls {
			if *acl.VpcId == *vpcId {
				_, err := client.DeleteNetworkAcl(&ec2.DeleteNetworkAclInput{NetworkAclId: acl.NetworkAclId})
				if err != nil {
					fmt.Println("ERROR : cant delete ACL ", *acl.NetworkAclId)
					fmt.Println(err)

				}
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

func DeleteSecurityGroups(client clientpkg.Client, vpcId *string) error {

	var token *string

	for {
		securityGroupList, err := client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{NextToken: token})
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR: Unable to retreive Security group list ")
			return err
		}

		for _, securityGroup := range securityGroupList.SecurityGroups {
			if *securityGroup.VpcId == *vpcId {
				_, err = client.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{IpPermissions: securityGroup.IpPermissions, GroupId: securityGroup.GroupId})
				if err != nil {
					fmt.Println("Some permissions have not been deleted")
					fmt.Println(err)
				}

			}
		}

		for _, securityGroup := range securityGroupList.SecurityGroups {
			if *securityGroup.VpcId == *vpcId {
				_, err = client.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{GroupId: securityGroup.GroupId})
				if err != nil {
					fmt.Println("ERROR : cant delete security group ", *securityGroup.GroupId)
					fmt.Println(err)

				}
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

func DeleteVpcEndpoint(client clientpkg.Client, vpcId *string) error {

	var vpcEndpointToBeDeleted []*string
	var token *string
	for {
		vpcEndpointList, err := client.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{NextToken: token})
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR: Unable to retreive VPC endpoint list ")
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
		fmt.Println("Following VPC could not be deleted ", vpcNotDeleted.String())
		fmt.Println(err)

	} else {
		fmt.Println("ALL VPC have been deleted successfully")

	}
	return nil
}

func DeleteVpnConnections(client clientpkg.Client) error {

	// does not require pagination
	vpnConnectionList, err := client.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{})
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR: Unable to retreive VPN connection list ")
		return err
	}

	for _, vpnConnection := range vpnConnectionList.VpnConnections {

		_, err = client.DeleteVpnConnection(&ec2.DeleteVpnConnectionInput{VpnConnectionId: vpnConnection.VpnConnectionId})
		if err != nil {
			fmt.Println("ERROR : cant delete VPN connection ", *vpnConnection.VpnConnectionId)
			fmt.Println(err)

		}

	}

	return nil

}

func DetachVpnGateway(client clientpkg.Client, vpcId *string) error {

	// does not require pagination
	vpnGatewayList, err := client.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{})
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR: Unable to retreive VPN gateway list ")
		return err
	}

	for _, vpnGateway := range vpnGatewayList.VpnGateways {

		_, err = client.DetachVpnGateway(&ec2.DetachVpnGatewayInput{VpcId: vpcId, VpnGatewayId: vpnGateway.VpnGatewayId})
		if err != nil {
			fmt.Println("ERROR : cant detach VPN gateway ", *vpnGateway.VpnGatewayId)
			fmt.Println(err)

		}

	}

	return nil

}
