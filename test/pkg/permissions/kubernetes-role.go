package permissions

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	typedRBACV1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesRole struct {
	roles           typedRBACV1.RoleInterface
	role            *rbacV1.Role
	serviceAccounts typedCoreV1.ServiceAccountInterface
	serviceAccount  *coreV1.ServiceAccount
	roleBindings    typedRBACV1.RoleBindingInterface
	roleBinding     *rbacV1.RoleBinding
}

func NewKubernetesRole(name, kubeConfigPath string,
	options ...kubernetesRoleOption,
) (
	r *KubernetesRole, e error,
) {
	const (
		bindingRoleRefAPIGroup = "rbac.authorization.k8s.io"
		bindingRoleRefKind     = "Role"
		bindingSubjectKind     = "ServiceAccount"
		masterURL              = ""
	)

	var (
		clientset *kubernetes.Clientset
		config    *rest.Config
		option    kubernetesRoleOption
	)

	config, e = clientcmd.BuildConfigFromFlags(masterURL, kubeConfigPath)
	if e != nil {
		return
	}

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		return
	}

	r = &KubernetesRole{
		roles: clientset.RbacV1().Roles(coreV1.NamespaceDefault),
		role: &rbacV1.Role{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
			Rules: make([]rbacV1.PolicyRule, 0),
		},
		serviceAccounts: clientset.CoreV1().ServiceAccounts(
			coreV1.NamespaceDefault,
		),
		serviceAccount: &coreV1.ServiceAccount{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
		},
		roleBindings: clientset.RbacV1().RoleBindings(coreV1.NamespaceDefault),
		roleBinding: &rbacV1.RoleBinding{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
			Subjects: []rbacV1.Subject{
				{
					Kind: bindingSubjectKind,
					Name: name,
				},
			},
			RoleRef: rbacV1.RoleRef{
				APIGroup: bindingRoleRefAPIGroup,
				Kind:     bindingRoleRefKind,
				Name:     name,
			},
		},
	}

	for _, option = range options {
		e = option(r)
		if e != nil {
			return
		}
	}

	r.role, e = r.roles.Create(
		context.Background(),
		r.role,
		metaV1.CreateOptions{},
	)
	if e != nil {
		return
	}

	r.serviceAccount, e = r.serviceAccounts.Create(
		context.Background(),
		r.serviceAccount,
		metaV1.CreateOptions{},
	)
	if e != nil {
		return
	}

	r.roleBinding, e = r.roleBindings.Create(
		context.Background(),
		r.roleBinding,
		metaV1.CreateOptions{},
	)
	if e != nil {
		return
	}

	return
}

func (r *KubernetesRole) Destroy() (e error) {
	e = r.roles.Delete(
		context.Background(),
		r.role.GetObjectMeta().GetName(),
		metaV1.DeleteOptions{},
	)
	if e != nil {
		return
	}

	e = r.serviceAccounts.Delete(
		context.Background(),
		r.serviceAccount.GetObjectMeta().GetName(),
		metaV1.DeleteOptions{},
	)
	if e != nil {
		return
	}

	e = r.roleBindings.Delete(
		context.Background(),
		r.roleBinding.GetObjectMeta().GetName(),
		metaV1.DeleteOptions{},
	)
	if e != nil {
		return
	}

	return
}

type kubernetesRoleOption func(*KubernetesRole) error

func WithPolicyRule(verbs, apiGroups, resources []string) (
	option kubernetesRoleOption,
) {
	option = func(r *KubernetesRole) (e error) {
		r.role.Rules = append(r.role.Rules,
			rbacV1.PolicyRule{
				Verbs:     verbs,
				APIGroups: apiGroups,
				Resources: resources,
			},
		)

		return
	}

	return
}
