package awsManager

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/go-logr/logr"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"github.com/openshift/aws-account-shredder/pkg/localMetrics"
)

// source : https://github.com/openshift/aws-account-operator/blob/master/pkg/controller/accountclaim/reuse.go#L321
// CleanUpAwsRoute53 cleans up awsRoute53
func CleanUpAwsRoute53(client clientpkg.Client, logger logr.Logger) error {

	var nextZoneMarker *string
	var errFlag bool = false

	// Paginate through hosted zones
	for {
		// Get list of hosted zones by page
		hostedZonesOutput, err := client.ListHostedZones(&route53.ListHostedZonesInput{Marker: nextZoneMarker})
		if err != nil {
			logger.Error(err, "Failed to retrieve hosted zones")
			// have to return here, or else invalid pointer reference will occur
			return err
		}

		for _, zone := range hostedZonesOutput.HostedZones {

			// List and delete all Record Sets for the current zone
			var nextRecordName *string
			// Pagination again!!!!!
			for {
				recordSet, listRecordsError := client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{HostedZoneId: zone.Id, StartRecordName: nextRecordName})
				if listRecordsError != nil {
					logger.Error(listRecordsError, "Failed to list Record sets for hosted zone", "Name", *zone.Name)
					errFlag = true
				}

				changeBatch := &route53.ChangeBatch{}
				for _, record := range recordSet.ResourceRecordSets {
					// Build ChangeBatch
					// https://docs.aws.amazon.com/sdk-for-go/api/service/route53/#ChangeBatch
					//https://docs.aws.amazon.com/sdk-for-go/api/service/route53/#Change
					if *record.Type != "NS" && *record.Type != "SOA" {
						changeBatch.Changes = append(changeBatch.Changes, &route53.Change{
							Action:            aws.String("DELETE"),
							ResourceRecordSet: record,
						})
					}
				}

				if changeBatch.Changes != nil {
					_, changeErr := client.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{HostedZoneId: zone.Id, ChangeBatch: changeBatch})
					if changeErr != nil {
						logger.Error(changeErr, "Failed to delete record sets for hosted zone", "Name", *zone.Name)
						errFlag = true
						localMetrics.ResourceFail(localMetrics.Route53RecordSet, client.GetRegion())
					} else {
						localMetrics.ResourceSuccess(localMetrics.Route53RecordSet, client.GetRegion())
					}
				}

				if *recordSet.IsTruncated {
					nextRecordName = recordSet.NextRecordName
				} else {
					break
				}

			}

			_, deleteError := client.DeleteHostedZone(&route53.DeleteHostedZoneInput{Id: zone.Id})
			if deleteError != nil {
				logger.Error(err, "failed to delete HostedZone", "ID", zone.Id)
				errFlag = true
				localMetrics.ResourceFail(localMetrics.Route53HostedZone, client.GetRegion())
				continue
			}
			localMetrics.ResourceSuccess(localMetrics.Route53HostedZone, client.GetRegion())
		}

		if *hostedZonesOutput.IsTruncated {
			nextZoneMarker = hostedZonesOutput.Marker
		} else {
			break
		}
	}

	// errFlag initially set to false
	if errFlag {
		return errors.New("ERROR")
	} else {
		logger.Info("Route53 cleanup finished successfully")
		return nil
	}
}
