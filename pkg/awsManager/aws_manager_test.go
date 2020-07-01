package awsManager

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"testing"

	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/mock/gomock"
	"github.com/openshift/aws-account-shredder/pkg/mock"
)

type mockSuite struct {
	mockCtrl      *gomock.Controller
	mockAWSClient *mock.MockClient
}

// setupDefaultMocks is an easy way to setup all of the default mocks
func setupDefaultMocks(t *testing.T) *mockSuite {
	mocks := &mockSuite{
		mockCtrl: gomock.NewController(t),
	}

	mocks.mockAWSClient = mock.NewMockClient(mocks.mockCtrl)
	return mocks
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

			},
			listOfBuckets: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected: true,
		}, {
			title: "test 3 - valid Buckets passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.BatchDeleteBucketObjects(gomock.Any()).Return(nil).AnyTimes()
				r.DeleteBucket(gomock.Any()).Return(&s3.DeleteBucketOutput{}, nil).AnyTimes()
			},
			listOfBuckets: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteS3Buckets(mocks.mockAWSClient, tc.listOfBuckets)

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
			},
			listOfInstances: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:   true,
		}, {
			title: "test 3 - valid Instances passed",
			setupAWSMock: func(r *mock.MockClientMockRecorder) {
				r.TerminateInstances(gomock.Any()).Return(&ec2.TerminateInstancesOutput{}, nil).AnyTimes()
			},
			listOfInstances: []*string{aws.String("abcd"), aws.String("abcd")},
			errorExpected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupDefaultMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			mockExecution := DeleteEc2Instance(mocks.mockAWSClient, tc.listOfInstances)

			if mockExecution != nil && tc.errorExpected == false {
				t.Errorf(tc.title, "Failed")
			}
		})
	}
}
