package pod

import (
	"sort"
	"testing"
	"time"

	resourceCommon "github.com/Netflix/titus-kube-common/resource"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ptr "k8s.io/utils/pointer"
)

func durationPtr(val string) *time.Duration {
	durVal, _ := time.ParseDuration(val)
	return &durVal
}

func stringToResourcePtr(val string) *resource.Quantity {
	resVal, _ := resource.ParseQuantity(val)
	return &resVal
}

func uint32Ptr(val uint32) *uint32 {
	ptrVal := &val
	return ptrVal
}

func uint64Ptr(val uint64) *uint64 {
	ptrVal := &val
	return ptrVal
}

func buildPod(annotations, labels map[string]string) *corev1.Pod {
	cpu := resource.NewQuantity(1, resource.DecimalSI)
	gpu := resource.NewQuantity(0, resource.DecimalSI)
	mem, _ := resource.ParseQuantity("512Mi")
	disk, _ := resource.ParseQuantity("10Gi")
	network, _ := resource.ParseQuantity("128M")

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "foo",
			Namespace:   "default",
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "task-id-in-container",
					Image: "my-registry.example.com/sample/helloworld:latest",
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:                 *cpu,
							corev1.ResourceMemory:              mem,
							corev1.ResourceEphemeralStorage:    disk,
							resourceCommon.ResourceNameGpu:     *gpu,
							resourceCommon.ResourceNameNetwork: network,
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:                 *cpu,
							corev1.ResourceMemory:              mem,
							corev1.ResourceEphemeralStorage:    disk,
							resourceCommon.ResourceNameGpu:     *gpu,
							resourceCommon.ResourceNameNetwork: network,
						},
					},
					TTY: true,
				},
			},
		},
	}

	return pod
}

func TestParsePod(t *testing.T) {
	taskId := "task-id-in-container"
	annotations := map[string]string{
		// strings
		AnnotationKeyPrefixAppArmor + "/" + taskId: "localhost/docker_titus",
		AnnotationKeyIAMRole:                       "arn:aws:iam::0:role/DefaultContainerRole",
		AnnotationKeyJobID:                         "myjobid",
		AnnotationKeyJobType:                       "BATCH",
		AnnotationKeyJobDescriptor:                 "myjobdesc",
		AnnotationKeyPodTitusContainerInfo:         "cinfo",
		AnnotationKeyWorkloadDetail:                "mydetail",
		AnnotationKeyWorkloadName:                  "myapp",
		AnnotationKeyWorkloadOwnerEmail:            "test@example.com",
		AnnotationKeyWorkloadSequence:              "v000",
		AnnotationKeyWorkloadStack:                 "mystack",

		AnnotationKeyNetworkAccountID:        "123456",
		AnnotationKeyNetworkElasticIPPool:    "pool-1",
		AnnotationKeyNetworkElasticIPs:       "eip-1,eip-2",
		AnnotationKeyNetworkIMDSRequireToken: "require-token",
		AnnotationKeyNetworkMode:             "example-network-mode",
		// Spaces intentionally added: we need to trim these
		AnnotationKeyNetworkSecurityGroups:         "sg-1 , sg-2 ",
		AnnotationKeyNetworkStaticIPAllocationUUID: "static-ip-alloc-id",
		AnnotationKeyNetworkSubnetIDs:              "subnet-1 , subnet-2 ",
		AnnotationKeyPodTitusSystemEnvVarNames:     "SYSTEM1 , SYSTEM2 ",

		// We don't parse these right now - including them so that
		// tests fail if we do start parsing them or remove them
		AnnotationKeyOpportunisticCPU:              "4",
		AnnotationKeyOpportunisticResourceID:       "op-res-id",
		AnnotationKeyPredictionRuntime:             "44",
		AnnotationKeyPredictionConfidence:          "5",
		AnnotationKeyPredictionModelID:             "model-id",
		AnnotationKeyPredictionModelVersion:        "v2",
		AnnotationKeyPredictionABTestCell:          "cell1",
		AnnotationKeyPredictionPredictionAvailable: "a,b",
		AnnotationKeyPredictionSelectorInfo:        "prediction",

		AnnotationKeySecurityWorkloadMetadata:    "app-metadata",
		AnnotationKeySecurityWorkloadMetadataSig: "app-metadata-sig",

		AnnotationKeyPodHostnameStyle: "ec2",
		AnnotationKeyPodSchedPolicy:   "batch",

		AnnotationKeyLogS3BucketName:    "bucket-name",
		AnnotationKeyLogS3PathPrefix:    "s3-prefix",
		AnnotationKeyLogS3WriterIAMRole: "arn:aws:iam::0:role/LogWriterRole",

		// bools
		AnnotationKeyLogKeepLocalFile:                 "true",
		AnnotationKeyNetworkAssignIPv6Address:         "true",
		AnnotationKeyNetworkBurstingEnabled:           "true",
		AnnotationKeyNetworkJumboFramesEnabled:        "true",
		AnnotationKeyPodCPUBurstingEnabled:            "true",
		AnnotationKeyPodFuseEnabled:                   "true",
		AnnotationKeyPodKvmEnabled:                    "true",
		AnnotationKeyPodSeccompAgentNetEnabled:        "true",
		AnnotationKeyPodSeccompAgentPerfEnabled:       "true",
		AnnotationKeyPodTitusEntrypointShellSplitting: "true",

		// ints
		AnnotationKeyPodSchemaVersion:       "2",
		AnnotationKeyJobAcceptedTimestampMs: "1602201163007",
		AnnotationKeyPodOomScoreAdj:         "-800",

		// resource values
		AnnotationKeyEgressBandwidth:  "10M",
		AnnotationKeyIngressBandwidth: "20M",

		// durations
		AnnotationKeyLogStdioCheckInterval:  "2m",
		AnnotationKeyLogUploadCheckInterval: "1m",
		AnnotationKeyLogUploadThresholdTime: "3m",

		// service config
		AnnotationKeyServicePrefix + "/servicemesh.v2.enabled": "true",
		AnnotationKeyServicePrefix + "/servicemesh.v2.image":   "titusops/servicemesh:latest",
	}

	labels := map[string]string{
		LabelKeyByteUnitsEnabled: "true",
		LabelKeyCapacityGroup:    "DEFAULT",
		LabelKeyTaskId:           "task-id-in-label",
	}

	pod := buildPod(annotations, labels)
	conf, err := PodToConfig(pod)
	assert.NilError(t, err)
	sgIDs := []string{"sg-1", "sg-2"}
	subnetIDs := []string{"subnet-1", "subnet-2"}
	expConf := Config{
		AppArmorProfile:          ptr.StringPtr("localhost/docker_titus"),
		AccountID:                ptr.StringPtr("123456"),
		WorkloadDetail:           ptr.StringPtr("mydetail"),
		WorkloadMetadata:         ptr.StringPtr("app-metadata"),
		WorkloadMetadataSig:      ptr.StringPtr("app-metadata-sig"),
		WorkloadName:             ptr.StringPtr("myapp"),
		WorkloadOwnerEmail:       ptr.StringPtr("test@example.com"),
		WorkloadSequence:         ptr.StringPtr("v000"),
		WorkloadStack:            ptr.StringPtr("mystack"),
		AssignIPv6Address:        ptr.BoolPtr(true),
		BytesEnabled:             ptr.BoolPtr(true),
		CapacityGroup:            ptr.StringPtr("DEFAULT"),
		ContainerInfo:            ptr.StringPtr("cinfo"),
		CPUBurstingEnabled:       ptr.BoolPtr(true),
		EgressBandwidth:          stringToResourcePtr("10M"),
		ElasticIPPool:            ptr.StringPtr("pool-1"),
		ElasticIPs:               ptr.StringPtr("eip-1,eip-2"),
		EntrypointShellSplitting: ptr.BoolPtr(true),
		FuseEnabled:              ptr.BoolPtr(true),
		HostnameStyle:            ptr.StringPtr("ec2"),
		IAMRole:                  ptr.StringPtr("arn:aws:iam::0:role/DefaultContainerRole"),
		IMDSRequireToken:         ptr.StringPtr("require-token"),
		IngressBandwidth:         stringToResourcePtr("20M"),
		JobAcceptedTimestampMs:   uint64Ptr(1602201163007),
		JobDescriptor:            ptr.StringPtr("myjobdesc"),
		JobID:                    ptr.StringPtr("myjobid"),
		JobType:                  ptr.StringPtr("BATCH"),
		JumboFramesEnabled:       ptr.BoolPtr(true),
		KvmEnabled:               ptr.BoolPtr(true),
		LogKeepLocalFile:         ptr.BoolPtr(true),
		LogStdioCheckInterval:    durationPtr("2m"),
		LogUploadCheckInterval:   durationPtr("1m"),
		LogUploadThresholdTime:   durationPtr("3m"),
		LogS3BucketName:          ptr.StringPtr("bucket-name"),
		LogS3PathPrefix:          ptr.StringPtr("s3-prefix"),
		LogS3WriterIAMRole:       ptr.StringPtr("arn:aws:iam::0:role/LogWriterRole"),
		NetworkMode:              ptr.StringPtr("example-network-mode"),
		NetworkBurstingEnabled:   ptr.BoolPtr(true),
		OomScoreAdj:              ptr.Int32Ptr(-800),
		PodSchemaVersion:         uint32Ptr(2),
		ResourceCPU:              stringToResourcePtr("1"),
		ResourceDisk:             stringToResourcePtr("10737418240"),
		ResourceMemory:           stringToResourcePtr("536870912"),
		ResourceNetwork:          stringToResourcePtr("128M"),
		ResourceGPU:              stringToResourcePtr("0"),
		SchedPolicy:              ptr.StringPtr("batch"),
		SeccompAgentNetEnabled:   ptr.BoolPtr(true),
		SeccompAgentPerfEnabled:  ptr.BoolPtr(true),
		SecurityGroupIDs:         &sgIDs,
		Sidecars: []Sidecar{
			{Name: "servicemesh", Enabled: true, Image: "titusops/servicemesh:latest", Version: 2},
		},
		StaticIPAllocationUUID: ptr.StringPtr("static-ip-alloc-id"),
		SubnetIDs:              &subnetIDs,
		SystemEnvVarNames:      []string{"SYSTEM1", "SYSTEM2"},
		TaskID:                 ptr.StringPtr("task-id-in-label"),
		TTYEnabled:             ptr.BoolPtr(true),
	}
	assert.DeepEqual(t, expConf, *conf)
}

func TestParsePodInvalid(t *testing.T) {
	badAnnotations := []struct {
		annotations map[string]string
		errMatch    string
	}{
		{
			annotations: map[string]string{
				AnnotationKeyPodHostnameStyle: "not-ec2",
			},
			errMatch: "annotation is not a valid hostname style: " + AnnotationKeyPodHostnameStyle,
		},
		{
			annotations: map[string]string{
				AnnotationKeyLogKeepLocalFile: "yes",
			},
			errMatch: "annotation is not a valid boolean value: " + AnnotationKeyLogKeepLocalFile,
		},
		{
			annotations: map[string]string{
				AnnotationKeyPodSchemaVersion: "-2",
			},
			errMatch: "annotation is not a valid uint32 value: " + AnnotationKeyPodSchemaVersion,
		},
		{
			annotations: map[string]string{
				AnnotationKeyJobAcceptedTimestampMs: "-5",
			},
			errMatch: "annotation is not a valid uint64 value: " + AnnotationKeyJobAcceptedTimestampMs,
		},
		{
			annotations: map[string]string{
				AnnotationKeyPodOomScoreAdj: "foo",
			},
			errMatch: "annotation is not a valid int32 value: " + AnnotationKeyPodOomScoreAdj,
		},
		{
			annotations: map[string]string{
				AnnotationKeyEgressBandwidth: "10ZiB",
			},
			errMatch: "annotation is not a valid resource value: " + AnnotationKeyEgressBandwidth,
		},
		{
			annotations: map[string]string{
				AnnotationKeyLogStdioCheckInterval: "2yearz",
			},
			errMatch: "annotation is not a valid duration value: " + AnnotationKeyLogStdioCheckInterval,
		},
		{
			annotations: map[string]string{
				AnnotationKeyPodSchedPolicy: "something",
			},
			errMatch: "annotation is not a valid scheduler policy: " + AnnotationKeyPodSchedPolicy,
		},
		{
			annotations: map[string]string{
				AnnotationKeyServicePrefix + "/only-one-field": "true",
			},
			errMatch: "annotation has an incorrect number of service configuration parameters: service.netflix.com/only-one-field",
		},
		{
			annotations: map[string]string{
				AnnotationKeyServicePrefix + "/foo.vA.enabled": "true",
			},
			errMatch: "annotation has an incorrect service version number: service.netflix.com/foo.vA.enabled",
		},
		{
			annotations: map[string]string{
				AnnotationKeyServicePrefix + "/foo.v1.enabled": "asdf",
			},
			errMatch: "annotation has an incorrect service enabled boolean value: service.netflix.com/foo.v1.enabled",
		},
		{
			annotations: map[string]string{
				AnnotationKeyServicePrefix + "/foo.v1.image": "asdf",
			},
			errMatch: "error parsing service image annotation: service.netflix.com/foo.v1.image: image does not have a digest or tag",
		},
	}

	for _, ann := range badAnnotations {
		pod := buildPod(ann.annotations, map[string]string{})
		_, err := PodToConfig(pod)
		assert.ErrorContains(t, err, ann.errMatch)
	}

	pod := buildPod(map[string]string{}, map[string]string{
		LabelKeyByteUnitsEnabled: "yep",
	})
	_, err := PodToConfig(pod)
	assert.ErrorContains(t, err, "label is not a valid boolean value: "+LabelKeyByteUnitsEnabled)
}

func TestBadBoolAnnotations(t *testing.T) {
	boolAnnotations := []string{
		AnnotationKeyLogKeepLocalFile,
		AnnotationKeyNetworkAssignIPv6Address,
		AnnotationKeyNetworkBurstingEnabled,
		AnnotationKeyNetworkJumboFramesEnabled,
		AnnotationKeyPodCPUBurstingEnabled,
		AnnotationKeyPodFuseEnabled,
		AnnotationKeyPodKvmEnabled,
		AnnotationKeyPodSeccompAgentNetEnabled,
		AnnotationKeyPodSeccompAgentPerfEnabled,
		AnnotationKeyPodTitusEntrypointShellSplitting,
	}

	for _, ann := range boolAnnotations {
		pod := buildPod(map[string]string{ann: "bad"}, map[string]string{})
		_, err := PodToConfig(pod)
		assert.ErrorContains(t, err, "annotation is not a valid boolean value: "+ann)
	}
}

func TestLogUploadRegExp(t *testing.T) {
	// You can't DeepEqual regexps, so test it separately
	annotations := map[string]string{
		AnnotationKeyLogUploadRegexp: ".*.foo",
	}
	labels := map[string]string{}

	pod := buildPod(annotations, labels)
	conf, err := PodToConfig(pod)
	assert.NilError(t, err)

	assert.Assert(t, conf.LogUploadRegExp != nil)
	assert.Equal(t, conf.LogUploadRegExp.String(), ".*.foo")
}

func TestServiceAnnotations(t *testing.T) {
	imgWithSha := "titusops/svc@sha256:5abd793cc69018e747cb8d4bc288f1df7b20747f91ec26da88f0fa4ba2ec46a1"
	annotations := map[string]string{
		AnnotationKeyServicePrefix + "/svc.v2.enabled": "false",
		AnnotationKeyServicePrefix + "/svc.v2.image":   "titusops/svc:latest",
		AnnotationKeyServicePrefix + "/svc.v3.enabled": "true",
		AnnotationKeyServicePrefix + "/svc.v3.image":   imgWithSha,
	}
	labels := map[string]string{}

	pod := buildPod(annotations, labels)
	conf, err := PodToConfig(pod)
	assert.NilError(t, err)

	assert.Assert(t, len(conf.Sidecars) == 2)
	sidecars := conf.Sidecars
	sort.Slice(sidecars, func(i, j int) bool {
		if sidecars[i].Name == sidecars[j].Name {
			return sidecars[i].Version < sidecars[j].Version
		}
		return sidecars[i].Name < sidecars[j].Name
	})

	assert.DeepEqual(t, []Sidecar{
		{Name: "svc", Enabled: false, Image: "titusops/svc:latest", Version: 2},
		{Name: "svc", Enabled: true, Image: imgWithSha, Version: 3},
	}, sidecars)
}

// XXX: test all nil
// XXX: test resources when bytes enabled
