#!/bin/bash

#function add-vm-to-mci() {

	echo "####################################################################"
	echo "## 9. Create vm on MCI"
	echo "####################################################################"

	source ../init.sh

	NUMVM=${OPTION01:-1}

	if [[ -z "${DISK_TYPE[$INDEX,$REGION]}" ]]; then
		RootDiskType="default"
	else
		RootDiskType="${DISK_TYPE[$INDEX,$REGION]}"
	fi

	if [[ -z "${DISK_SIZE[$INDEX,$REGION]}" ]]; then
		RootDiskSize="default"
	else
		RootDiskSize="${DISK_SIZE[$INDEX,$REGION]}"
	fi


	# Get a Subnet from the vNet (10.25.10.1/24 condition is only for KT cloud vpc)
	echo "- Get vNet ${CONN_CONFIG[$INDEX,$REGION]}-${POSTFIX} to designate a Subnet"
	VNETINFO=$(curl -H "${AUTH}" -sX GET http://$TumblebugServer/tumblebug/ns/$NSID/resources/vNet/${CONN_CONFIG[$INDEX,$REGION]}-${POSTFIX})
	SUBNETARRAY=$(jq -r '.subnetInfoList' <<<"$VNETINFO")
	echo "$SUBNETARRAY"
	SUBNETID=$(jq -r '.subnetInfoList[] | select(.IPv4_CIDR == "10.25.10.1/24") | .Id' <<<"$VNETINFO")
	if [ -z "$SUBNETID" ]; then
		SUBNETID=$(jq -r '.subnetInfoList[0].Id' <<<"$VNETINFO")
	fi
	echo "Designated Subnet ID (for testing only): $SUBNETID"

	
	curl -H "${AUTH}" -sX POST http://$TumblebugServer/tumblebug/ns/$NSID/mci/$MCIID/vm -H 'Content-Type: application/json' -d \
		'{
			"subGroupSize": "'${NUMVM}'",
			"name": "'${CONN_CONFIG[$INDEX,$REGION]}'",
			"imageId": "'${CONN_CONFIG[$INDEX,$REGION]}'-'${POSTFIX}'",
			"vmUserName": "cb-user",
			"connectionName": "'${CONN_CONFIG[$INDEX,$REGION]}'",
			"sshKeyId": "'${CONN_CONFIG[$INDEX,$REGION]}'-'${POSTFIX}'",
			"specId": "'${CONN_CONFIG[$INDEX,$REGION]}'-'${POSTFIX}'",
			"securityGroupIds": [
				"'${CONN_CONFIG[$INDEX,$REGION]}'-'${POSTFIX}'"
			],
			"vNetId": "'${CONN_CONFIG[$INDEX,$REGION]}'-'${POSTFIX}'",
			"subnetId": "'${SUBNETID}'",
			"description": "description",
			"vmUserPassword": "",
			"rootDiskType": "'${RootDiskType}'",
			"rootDiskSize": "'${RootDiskSize}'"
		}' | jq '.' 
#}

#add-vm-to-mci