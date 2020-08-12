### Troubleshooting / Debugging the Undeleted Resources


AWS resources that can be deleted by the shredder are the following: 

````
EC2 Instances
S3 Buckets
Route 53 resources
VPC Instances and endpoints
Elastic Load Balancer
NAT Gateway
Network Load Balancer
Network Interface
Internet Gateway
Subnets
Route Tables
Network ACL 
Security Groups
VPN gateway and endpoints
Elastic File storage mount targets and volumes
````

In case additional resources need to be deleted, the logic for that has to be programmed in the directory `````/pkg/awsManager`````


If any of the above resources are not deleted, it can be because of 2 main reasons : 

1) It is the 'Default' type. Like, a default route or default ACL. Some default resources can not be deleted. Default resource do not block the deletion of any Resource. But they may require some pre-processing.
Example: A default routing table can not be deleted, but it has to be cleaned up such that all the routes get deleted. 

2) The resource is blocked by some other resource. For example : A VPC can not be deleted until and unless all the resources in that VPC has been removed. 


It might take some time  for the resources to delete. Checking the shredder logs will give more information on why the resource was not deleted.

While development / testing , I found that there are a lot of dependencies for VPC and security group. 
I have included some script that will list those dependencies


### Details about files in this directory

##### security_group_blocked_resources.sh

This script will give a list of all the resources that are blocking a particular security group from being deleted.

Replace the group-id with the security group id of the group. Also, remove the <> brackets. 

Not every resource listed by the script need to be deleted ( some of them will be 'default' resource). However , if a security group is not deleted after several pass , this might give a better insight on the issue.

##### vpc_blocked_resources.sh

This script will list all the resources that are blocking a particular VPC from being deleted

Replace the vpc-xxxxxxxxxxxxx with the vpc-id

Not every resource listed by the script need to be deleted ( some of them will be 'default' resource). However , if a VPC is not deleted after several pass , this might give a better insight on the issue.
