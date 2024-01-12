// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/build/pkg/config"
	"github.com/shipwright-io/build/pkg/env"
	"github.com/shipwright-io/build/pkg/reconciler/buildrun/resources/steps"
	"github.com/shipwright-io/build/pkg/volumes"
	pipelineapi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
)

const (
	prefixParamsResultsVolumes = "shp"

	paramOutputImage    = "output-image"
	paramOutputInsecure = "output-insecure"
	paramSourceRoot     = "source-root"
	paramSourceContext  = "source-context"

	workspaceSource = "source"

	inputParamContextDir = "CONTEXT_DIR"
)

// getStringTransformations gets us MANDATORY replacements using
// a poor man's templating mechanism - TODO: Use golang templating
func getStringTransformations(fullText string) string {

	stringTransformations := map[string]string{
		// this will be removed, build strategy author should use $(params.shp-output-image) directly
		"$(build.output.image)": fmt.Sprintf("$(params.%s-%s)", prefixParamsResultsVolumes, paramOutputImage),

		// this will be removed, build strategy author should use $(params.shp-source-context); it is still needed by the ko build
		// strategy that mis-uses this setting to store the path to the main package; requires strategy parameter support to get rid
		"$(build.source.contextDir)": fmt.Sprintf("$(inputs.params.%s)", inputParamContextDir),
	}

	// Run the text through all possible replacements
	for k, v := range stringTransformations {
		fullText = strings.ReplaceAll(fullText, k, v)
	}
	return fullText
}

// GenerateTaskSpec creates Tekton TaskRun spec to be used for a build run
func GenerateTaskSpec(
	cfg *config.Config,
	build *buildv1beta1.Build,
	buildRun *buildv1beta1.BuildRun,
	buildSteps []buildv1beta1.Step,
	parameterDefinitions []buildv1beta1.Parameter,
	buildStrategyVolumes []buildv1beta1.BuildStrategyVolume,
) (*pipelineapi.TaskSpec, error) {
	generatedTaskSpec := pipelineapi.TaskSpec{
		Params: []pipelineapi.ParamSpec{
			{
				// CONTEXT_DIR comes from the git source specification
				// in the Build object
				Description: "The root of the code",
				Name:        inputParamContextDir,
				Default: &pipelineapi.ParamValue{
					Type:      pipelineapi.ParamTypeString,
					StringVal: ".",
				},
			},
			{
				Name:        fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramOutputImage),
				Description: "The URL of the image that the build produces",
				Type:        pipelineapi.ParamTypeString,
			},
			{
				Name:        fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramOutputInsecure),
				Description: "A flag indicating that the output image is on an insecure container registry",
				Type:        pipelineapi.ParamTypeString,
			},
			{
				Name:        fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramSourceContext),
				Description: "The context directory inside the source directory",
				Type:        pipelineapi.ParamTypeString,
			},
			{
				Name:        fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramSourceRoot),
				Description: "The source directory",
				Type:        pipelineapi.ParamTypeString,
			},
		},
		Workspaces: []pipelineapi.WorkspaceDeclaration{
			// workspace for the source files
			{
				Name: workspaceSource,
			},
		},
	}

	generatedTaskSpec.Results = append(getTaskSpecResults(), getFailureDetailsTaskSpecResults()...)

	// define the results, steps and volumes for sources, or alternatively, wait for user upload
	AmendTaskSpecWithSources(cfg, &generatedTaskSpec, build, buildRun)

	// Add the strategy defined parameters into the Task spec
	for _, parameterDefinition := range parameterDefinitions {

		param := pipelineapi.ParamSpec{
			Name:        parameterDefinition.Name,
			Description: parameterDefinition.Description,
		}

		switch parameterDefinition.Type {
		case "": // string is default
			fallthrough
		case buildv1beta1.ParameterTypeString:
			param.Type = pipelineapi.ParamTypeString
			if parameterDefinition.Default != nil {
				param.Default = &pipelineapi.ParamValue{
					Type:      pipelineapi.ParamTypeString,
					StringVal: *parameterDefinition.Default,
				}
			}

		case buildv1beta1.ParameterTypeArray:
			param.Type = pipelineapi.ParamTypeArray
			if parameterDefinition.Defaults != nil {
				param.Default = &pipelineapi.ParamValue{
					Type:     pipelineapi.ParamTypeArray,
					ArrayVal: *parameterDefinition.Defaults,
				}
			}
		}

		generatedTaskSpec.Params = append(generatedTaskSpec.Params, param)
	}

	// Combine the environment variables specified in the Build object and the BuildRun object
	// env vars in the BuildRun supercede those in the Build, overwriting them
	combinedEnvs, err := env.MergeEnvVars(buildRun.Spec.Env, build.Spec.Env, true)
	if err != nil {
		return nil, err
	}

	// This map will contain mapping from all volume mount names that build steps contain
	// to their readonly status in order to later quickly check whether mount is correct
	volumeMounts := make(map[string]bool)
	buildStrategyVolumesMap := toVolumeMap(buildStrategyVolumes)

	// define the steps coming from the build strategy
	for _, containerValue := range buildSteps {

		var taskCommand []string
		for _, buildStrategyCommandPart := range containerValue.Command {
			taskCommand = append(taskCommand, getStringTransformations(buildStrategyCommandPart))
		}

		var taskArgs []string
		for _, buildStrategyArgPart := range containerValue.Args {
			taskArgs = append(taskArgs, getStringTransformations(buildStrategyArgPart))
		}

		taskImage := getStringTransformations(containerValue.Image)

		// Any collision between the env vars in the Container step and those in the Build/BuildRun
		// will result in an error and cause a failed TaskRun
		stepEnv, err := env.MergeEnvVars(combinedEnvs, containerValue.Env, false)
		if err != nil {
			return &generatedTaskSpec, fmt.Errorf("error(s) occurred merging environment variables into BuildStrategy %q steps: %s", build.Spec.StrategyName(), err.Error())
		}

		step := pipelineapi.Step{
			Image:            taskImage,
			ImagePullPolicy:  containerValue.ImagePullPolicy,
			Name:             containerValue.Name,
			VolumeMounts:     containerValue.VolumeMounts,
			Command:          taskCommand,
			Args:             taskArgs,
			SecurityContext:  containerValue.SecurityContext,
			WorkingDir:       containerValue.WorkingDir,
			ComputeResources: containerValue.Resources,
			Env:              stepEnv,
		}

		generatedTaskSpec.Steps = append(generatedTaskSpec.Steps, step)

		for _, vm := range containerValue.VolumeMounts {
			// here we should check that volume actually exists for this mount
			// and in case it does not, exit early with an error
			if _, ok := buildStrategyVolumesMap[vm.Name]; !ok {
				return nil, fmt.Errorf("volume for the Volume Mount %q is not found", vm.Name)
			}
			volumeMounts[vm.Name] = vm.ReadOnly
		}
	}

	// Add volumes from the strategy to generated task spec
	volumes, err := volumes.TaskSpecVolumes(volumeMounts, buildStrategyVolumes, build.Spec.Volumes, buildRun.Spec.Volumes)
	if err != nil {
		return nil, err
	}
	generatedTaskSpec.Volumes = append(generatedTaskSpec.Volumes, volumes...)

	return &generatedTaskSpec, nil
}

// GenerateTaskRun creates a Tekton TaskRun to be used for a build run
func GenerateTaskRun(
	cfg *config.Config,
	build *buildv1beta1.Build,
	buildRun *buildv1beta1.BuildRun,
	serviceAccountName string,
	strategy buildv1beta1.BuilderStrategy,
) (*pipelineapi.TaskRun, error) {

	// retrieve expected imageURL form build or buildRun
	var image string
	if buildRun.Spec.Output != nil {
		image = buildRun.Spec.Output.Image
	} else {
		image = build.Spec.Output.Image
	}

	insecure := false
	if buildRun.Spec.Output != nil && buildRun.Spec.Output.Insecure != nil {
		insecure = *buildRun.Spec.Output.Insecure
	} else if build.Spec.Output.Insecure != nil {
		insecure = *build.Spec.Output.Insecure
	}

	taskSpec, err := GenerateTaskSpec(
		cfg,
		build,
		buildRun,
		strategy.GetBuildSteps(),
		strategy.GetParameters(),
		strategy.GetVolumes(),
	)
	if err != nil {
		return nil, err
	}

	// Add BuildRun name reference to the TaskRun labels
	taskRunLabels := map[string]string{
		buildv1beta1.LabelBuildRun:           buildRun.Name,
		buildv1beta1.LabelBuildRunGeneration: strconv.FormatInt(buildRun.Generation, 10),
	}

	// Add Build name reference unless it is an embedded Build (empty build name)
	if build.Name != "" {
		taskRunLabels[buildv1beta1.LabelBuild] = build.Name
		taskRunLabels[buildv1beta1.LabelBuildGeneration] = strconv.FormatInt(build.Generation, 10)
	}

	expectedTaskRun := &pipelineapi.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: buildRun.Name + "-",
			Namespace:    buildRun.Namespace,
			Labels:       taskRunLabels,
		},
		Spec: pipelineapi.TaskRunSpec{
			ServiceAccountName: serviceAccountName,
			TaskSpec:           taskSpec,
			Workspaces: []pipelineapi.WorkspaceBinding{
				// workspace for the source files
				{
					Name:     workspaceSource,
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	// assign the annotations from the build strategy, filter out those that should not be propagated
	taskRunAnnotations := make(map[string]string)
	for key, value := range strategy.GetAnnotations() {
		if isPropagatableAnnotation(key) {
			taskRunAnnotations[key] = value
		}
	}

	// Update the security context of the Shipwright-injected steps with the runAs user of the build strategy
	steps.UpdateSecurityContext(taskSpec, taskRunAnnotations, strategy.GetBuildSteps(), strategy.GetSecurityContext())

	if len(taskRunAnnotations) > 0 {
		expectedTaskRun.Annotations = taskRunAnnotations
	}

	for label, value := range strategy.GetResourceLabels() {
		expectedTaskRun.Labels[label] = value
	}

	expectedTaskRun.Spec.Timeout = effectiveTimeout(build, buildRun)

	params := []pipelineapi.Param{
		{
			// shp-output-image
			Name: fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramOutputImage),
			Value: pipelineapi.ParamValue{
				Type:      pipelineapi.ParamTypeString,
				StringVal: image,
			},
		},
		{
			// shp-output-insecure
			Name: fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramOutputInsecure),
			Value: pipelineapi.ParamValue{
				Type:      pipelineapi.ParamTypeString,
				StringVal: strconv.FormatBool(insecure),
			},
		},
		{
			Name: fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramSourceRoot),
			Value: pipelineapi.ParamValue{
				Type:      pipelineapi.ParamTypeString,
				StringVal: "/workspace/source",
			},
		},
	}

	if build.Spec.Source.ContextDir != nil {
		params = append(params, pipelineapi.Param{
			Name: inputParamContextDir,
			Value: pipelineapi.ParamValue{
				Type:      pipelineapi.ParamTypeString,
				StringVal: *build.Spec.Source.ContextDir,
			},
		})
		params = append(params, pipelineapi.Param{
			Name: fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramSourceContext),
			Value: pipelineapi.ParamValue{
				Type:      pipelineapi.ParamTypeString,
				StringVal: path.Join("/workspace/source", *build.Spec.Source.ContextDir),
			},
		})
	} else {
		params = append(params, pipelineapi.Param{
			Name: fmt.Sprintf("%s-%s", prefixParamsResultsVolumes, paramSourceContext),
			Value: pipelineapi.ParamValue{
				Type:      pipelineapi.ParamTypeString,
				StringVal: "/workspace/source",
			},
		})
	}

	expectedTaskRun.Spec.Params = params

	// Ensure a proper override of params between Build and BuildRun
	// A BuildRun can override a param as long as it was defined in the Build
	paramValues := OverrideParams(build.Spec.ParamValues, buildRun.Spec.ParamValues)

	// Append params to the TaskRun spec definition
	for _, paramValue := range paramValues {
		parameterDefinition := FindParameterByName(strategy.GetParameters(), paramValue.Name)
		if parameterDefinition == nil {
			// this error should never happen because we validate this upfront in ValidateBuildRunParameters
			return nil, fmt.Errorf("the parameter %q is not defined in the build strategy %q", paramValue.Name, strategy.GetName())
		}

		if err := HandleTaskRunParam(expectedTaskRun, parameterDefinition, paramValue); err != nil {
			return nil, err
		}
	}

	// Setup image processing, this can be a no-op if no annotations or labels need to be mutated,
	// and if the strategy is pushing the image by not using $(params.shp-output-directory)
	buildRunOutput := buildRun.Spec.Output
	if buildRunOutput == nil {
		buildRunOutput = &buildv1beta1.Image{}
	}
	SetupImageProcessing(expectedTaskRun, cfg, build.Spec.Output, *buildRunOutput)

	return expectedTaskRun, nil
}

func effectiveTimeout(build *buildv1beta1.Build, buildRun *buildv1beta1.BuildRun) *metav1.Duration {
	if buildRun.Spec.Timeout != nil {
		return buildRun.Spec.Timeout

	} else if build.Spec.Timeout != nil {
		return build.Spec.Timeout
	}

	return nil
}

// isPropagatableAnnotation filters the last-applied-configuration annotation from kubectl because this would break the meaning of this annotation on the target object;
// also, annotations using our own custom resource domains are filtered out because we have no annotations with a semantic for both TaskRun and Pod
func isPropagatableAnnotation(key string) bool {
	return key != "kubectl.kubernetes.io/last-applied-configuration" &&
		!strings.HasPrefix(key, buildv1beta1.ClusterBuildStrategyDomain+"/") &&
		!strings.HasPrefix(key, buildv1beta1.BuildStrategyDomain+"/") &&
		!strings.HasPrefix(key, buildv1beta1.BuildDomain+"/") &&
		!strings.HasPrefix(key, buildv1beta1.BuildRunDomain+"/")
}

func toVolumeMap(strategyVolumes []buildv1beta1.BuildStrategyVolume) map[string]bool {
	res := make(map[string]bool)
	for _, v := range strategyVolumes {
		res[v.Name] = true
	}
	return res
}
