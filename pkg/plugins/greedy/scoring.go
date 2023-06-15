package greedy

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const (
	greedyModeName = "GreedyMode"
)

type GreedyMode struct{}

func (g *GreedyMode) Name() string {
	return greedyModeName
}

func (g *GreedyMode) Score(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, node *corev1.Node) (int64, *framework.Status) {
	// 在这里实现评分逻辑
	// 根据节点上GPU的空闲内存大小给节点打分
	// 评分越高表示节点上GPU的空闲内存越大，适合调度Pod

	// 获取节点上的GPU资源信息
	gpuResources := node.Status.Capacity[corev1.ResourceName("nvidia.com/gpu")]
	if gpuResources.Value() == 0 {
		// 如果节点上没有GPU资源，则返回最低分
		return 0, nil
	}

	// 获取节点上已经运行的Pod列表
	podList, err := state.GetPodLister().Pods(pod.Namespace).List(labels.Everything())
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Failed to get pod list: %v", err))
	}

	// 计算节点评分，评分越高表示节点上GPU的空闲内存越大
	score := int64(0)
	for _, existingPod := range podList {
		if existingPod.Spec.NodeName == node.Name && hasGPUResource(existingPod) {
			// 计算已分配给该节点上其他Pod的GPU内存大小
			allocatedGPU := existingPod.Spec.Containers[0].Resources.Requests[corev1.ResourceName("nvidia.com/gpu")]
			score -= allocatedGPU.Value()
		}
	}

	// 添加节点上剩余的GPU内存大小作为额外的评分加分项
	score += gpuResources.Value()
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

func (g *GreedyMode) NormalizeScore(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
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
