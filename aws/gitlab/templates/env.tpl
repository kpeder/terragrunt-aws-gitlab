---
environment: ENVIRONMENT
labels:
  deployment: PREFIX
  environment: ENVIRONMENT
  owner: OWNER
  team: TEAM

dependencies:
  custom_vpc_dependency_path: "reg-primary/vpcs/custom"
  custom_vpc_mock_outputs:
    azs:
      - "PREGIONa"
      - "PREGIONb"
      - "PREGIONc"
    igw_id: "igw-01934b9031b6f7518"
    name: "PREFIX-ENVIRONMENT-custom"
    nat_public_ips:
      - "1.2.3.4"
      - "2.3.4.5"
      - "3.4.5.6"
    natgw_ids:
      - "nat-005afb3d3d531b4ac"
      - "nat-0d05f8e74f6e5b291"
      - "nat-02b0a9954c6b7b626"
    private_subnets:
      - "subnet-06aa014d10fd0f6db"
      - "subnet-074f51b17b837a76d"
      - "subnet-07a56fe66308f6d8e"
    private_subnets_cidr_blocks:
      - "172.16.12.0/22"
      - "172.16.16.0/22"
      - "172.16.20.0/22"
    public_subnets:
      - "subnet-003fa0c4735ef22b2"
      - "subnet-02c5f074a651ac191"
      - "subnet-00d887d100903d07f"
    public_subnets_cidr_blocks:
      - "172.16.4.0/24"
      - "172.16.5.0/24"
      - "172.16.6.0/24"
    vpc_cidr_block: "172.16.0.0/16"
    vpc_enable_dns_hostnames: true
    vpc_enable_dns_support: true
    vpc_id: "vpc-0d8148e657a7787f1"
    vpc_main_route_table_id: "rtb-0ade48517f021bfde"

  custom_ice_sg_dependency_path: "reg-primary/sgs/custom-ice"
  custom_ice_sg_mock_outputs:
    security_group_id: "sg-06e47f69"
    security_group_name: "PREFIX-ENVIRONMENT-custom-ice"
    security_group_vpc_id: "vpc-0d8148e657a7787f1"

  gitlab_keypair_dependency_path: "reg-primary/keypairs/gitlab"
  gitlab_keypair_mock_outputs:
    key_pair_id: "key-0576e69c4b8faacc2"
    key_pair_name: "PREFIX-ENVIRONMENT-gitlab"

  gitlab_sg_dependency_path: "reg-primary/sgs/gitlab"
  gitlab_sg_mock_outputs:
    security_group_id: "sg-03d25a67"
    security_group_name: "PREFIX-ENVIRONMENT-gitlab"
    security_group_vpc_id: "vpc-0d8148e657a7787f1"
