package binback

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const (
	packingModeName = "PackingMode"
)

type PackingMode struct{}

func (p *PackingMode) Name() string {
	return packingModeName
}

func (p *PackingMode) Score(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, node *corev1.Node) (int64, *framework.Status) {
	// 在这里实现评分逻辑
	// 根据一块GPU上已经运行的Pod数量，给当前节点打分
	// 评分越高表示节点上已经使用的GPU资源越少，适合调度更多的Pod

	// 假设每块GPU最大支持的Pod数量为maxPodsPerGPU
	maxPodsPerGPU := 8

	// 获取节点上已经运行的Pod列表
	podList, err := state.GetPodLister().Pods(pod.Namespace).List(labels.Everything())
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Failed to get pod list: %v", err))
	}

	// 统计节点上已经使用的GPU资源数量
	usedGPUCount := 0
	for _, existingPod := range podList {
		if existingPod.Spec.NodeName == node.Name && hasGPUResource(existingPod) {
			usedGPUCount++
		}
	}

	// 计算节点评分，评分越高表示节点上已经使用的GPU资源越少
	score := int64(maxPodsPerGPU - usedGPUCount)
	return score, nil
}

func hasGPUResource(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if quantity, found := container.Resources.Requests[corev1.ResourceName("nvidia.com/gpu")]; found && quantity.Value() > 0 {
			return true
		}
	}
	return false
}

func (p *PackingMode) NormalizeScore(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var (
		highest int64 = 0
		lowest        = scores[0].Score
	)

	for _, nodeScore := range scores {
		if nodeScore.Score < lowest {
			lowest = nodeScore.Score
		}
		if nodeScore.Score > highest {
			highest = nodeScore.Score
		}
	}

	if highest == lowest {
		lowest--
	}

	// Set Range to [0-100]
	for i, nodeScore := range scores {
		scores[i].Score = (nodeScore.Score - lowest) * framework.MaxNodeScore / (highest - lowest)
		klog.Infof("Node: %v, Score: %v in Plugin: Mandalorian When scheduling Pod: %v/%v", scores[i].Name, scores[i].Score, pod.GetNamespace(), pod.GetName())
	}
	return nil
}