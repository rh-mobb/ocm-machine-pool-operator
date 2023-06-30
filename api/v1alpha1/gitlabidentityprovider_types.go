/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"fmt"

	clustersmgmtv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rh-mobb/ocm-operator/pkg/kubernetes"
)

const (
	GitLabClientSecretKey = "clientSecret"
	GitLabAccessTokenKey  = "accessToken"
	GitLabCAKey           = "ca.crt"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitLabIdentityProviderSpec defines the desired state of GitLabIdentityProvider.
//
//nolint:lll
type GitLabIdentityProviderSpec struct {
	// clientID is the oauth client ID
	ClientID string `json:"clientID"`

	// clientSecret is a required reference to the secret by name containing the oauth client secret.
	// The key "clientSecret" is used to locate the data.
	// If the secret or expected key is not found, the identity provider is not honored.
	// This should exist in the same namespace as the operator.
	ClientSecret configv1.SecretNameReference `json:"clientSecret"`

	// url is the oauth server base URL
	URL string `json:"url"`

	// ca is an optional reference to a config map by name containing the PEM-encoded CA bundle.
	// It is used as a trust anchor to validate the TLS certificate presented by the remote server.
	// The key "ca.crt" is used to locate the data.
	// If specified and the config map or expected key is not found, the identity provider is not honored.
	// If the specified ca data is not valid, the identity provider is not honored.
	// If empty, the default system roots are used.
	// This should exist in the same namespace as the operator.
	// +optional
	CA configv1.ConfigMapNameReference `json:"ca"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=claim
	// +kubebuilder:validation:Enum=claim;lookup;generate;add
	// Mapping method to use for the identity provider.
	// See https://docs.openshift.com/container-platform/latest/authentication/understanding-identity-provider.html#identity-provider-parameters_understanding-identity-provider
	// for a detailed description of what these mean.  Must be one of claim (default), lookup, generate, or add.
	MappingMethod string `json:"mappingMethod,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:message="clusterName is immutable",rule=(self == oldSelf)
	// Cluster ID in OpenShift Cluster Manager by which this should be managed for.  The cluster ID
	// can be obtained on the Clusters page for the individual cluster.  It may also be known as the
	// 'External ID' in some CLI clients.  It shows up in the format of 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx'
	// where the 'x' represents any alphanumeric character.
	ClusterName string `json:"clusterName,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=4
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:XValidation:message="displayName is immutable",rule=(self == oldSelf)
	// Friendly display name as displayed in the OpenShift Cluster Manager
	// console.  If this is empty, the metadata.name field of the parent resource is used
	// to construct the display name.  This is limited to 15 characters as per the backend
	// API limitation.
	DisplayName string `json:"displayName,omitempty"`

	// TODO: eventually we want to be able to have the operator create the application.  currently there is a limitation
	//       in gitlab which restricts application creation for a particular group to the server admins.  once this
	//       api limitation is removed (if ever) we can implement the following and be able to reconcile accordingly.

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:message="accessTokenSecret is immutable",rule=(self == oldSelf)
	// accessTokenSecret is a required reference to the secret by
	// name containing the GitLab access token required to interact with
	// the GitLab API.  This access token must have read/write API access.  The
	// secret must contain the key 'accessToken' to locate the data. If the secret or
	// expected key is not found, the identity provider is not honored. The namespace
	// for this secret must exist in the same namespace as the resource.
	// AccessTokenSecret string `json:"accessTokenSecret,omitempty"`
}

// GitLabIdentityProviderStatus defines the observed state of GitLabIdentityProvider.
type GitLabIdentityProviderStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +kubebuilder:validation:XValidation:message="status.clusterID is immutable",rule=(self == oldSelf)
	// Represents the programmatic cluster ID of the cluster, as
	// determined during reconciliation.  This is used to reduce
	// the number of API calls to look up a cluster ID based on
	// the cluster name.
	ClusterID string `json:"clusterID,omitempty"`

	// +kubebuilder:validation:XValidation:message="status.providerID is immutable",rule=(self == oldSelf)
	// Represents the programmatic identity provider ID of the IDP, as
	// determined during reconciliation.  This is used to reduce
	// the number of API calls to look up a cluster ID based on
	// the identity provider name.
	ProviderID string `json:"providerID,omitempty"`

	// +kubebuilder:validation:XValidation:message="status.callbackURL is immutable",rule=(self == oldSelf)
	// Represents the OAuth endpoint used for the OAuth provider to call back
	// to.  This is necessary for proper configuration of any external identity provider.
	CallbackURL string `json:"callbackURL,omitempty"`
}

// +kubebuilder:resource:categories=idps;identityproviders
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:validation:XValidation:message="metadata.name limited to 15 characters",rule=(self.metadata.name.size() <= 15)

// GitLabIdentityProvider is the Schema for the gitlabidentityproviders API.
type GitLabIdentityProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitLabIdentityProviderSpec   `json:"spec,omitempty"`
	Status GitLabIdentityProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitLabIdentityProviderList contains a list of GitLabIdentityProvider.
type GitLabIdentityProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitLabIdentityProvider `json:"items"`
}

// FindAll gets a complete list of resources in the cluster for this type.
func (gitlab *GitLabIdentityProvider) FindAll(
	ctx context.Context,
	c kubernetes.Client,
) ([]GitLabIdentityProvider, error) {
	objects := &GitLabIdentityProviderList{}

	if err := c.List(ctx, objects); err != nil {
		return []GitLabIdentityProvider{}, fmt.Errorf("unable to retrieve gitlab identity providers - %w", err)
	}

	return objects.Items, nil
}

// FindAllByClusterID gets a list of resources which have a particular cluster ID in the status field.
func (gitlab *GitLabIdentityProvider) FindAllByClusterID(
	ctx context.Context,
	c kubernetes.Client,
	clusterID string,
) ([]*GitLabIdentityProvider, error) {
	objects, err := gitlab.FindAll(ctx, c)
	if err != nil {
		return []*GitLabIdentityProvider{}, err
	}

	matches := []*GitLabIdentityProvider{}

	for i := range objects {
		if objects[i].Status.ClusterID == clusterID {
			matches = append(matches, &objects[i])
		}
	}

	return matches, nil
}

// ExistsForClusterID returns if a particular object is associated with a cluster ID.
func (gitlab *GitLabIdentityProvider) ExistsForClusterID(
	ctx context.Context,
	c kubernetes.Client,
	clusterID string,
) (bool, error) {
	objects, err := gitlab.FindAllByClusterID(ctx, c, clusterID)

	return (len(objects) > 0), err
}

// GetClusterID gets the status.clusterID field from the object.  It is used to
// satisfy the Workload interface.
func (gitlab *GitLabIdentityProvider) GetClusterID() string {
	return gitlab.Status.ClusterID
}

// GetConditions returns the status.conditions field from the object.  It is used to
// satisfy the Workload interface.
func (gitlab *GitLabIdentityProvider) GetConditions() []metav1.Condition {
	return gitlab.Status.Conditions
}

// SetConditions sets the status.conditions field from the object.  It is used to
// satisfy the Workload interface.
func (gitlab *GitLabIdentityProvider) SetConditions(conditions []metav1.Condition) {
	gitlab.Status.Conditions = conditions
}

// CopyFrom copies a GitLab Identity provider into an object that is able to be reconciled.
func (gitlab *GitLabIdentityProvider) CopyFrom(source *clustersmgmtv1.IdentityProvider) {
	gitlab.Spec.CA = configv1.ConfigMapNameReference{Name: source.Gitlab().CA()}
	gitlab.Spec.URL = source.Gitlab().URL()
	gitlab.Spec.ClientID = source.Gitlab().ClientID()
}

// Builder returns the builder object from a reconciler object.  This object is used to
// pass into the OCM API for creating the object.
func (gitlab *GitLabIdentityProvider) Builder(ca, clientSecret string) *clustersmgmtv1.IdentityProviderBuilder {
	builder := clustersmgmtv1.NewIdentityProvider().
		MappingMethod(clustersmgmtv1.IdentityProviderMappingMethod(gitlab.Spec.MappingMethod)).
		Name(gitlab.Spec.DisplayName).
		Type(clustersmgmtv1.IdentityProviderTypeGitlab)

	gitlabIDP := clustersmgmtv1.NewGitlabIdentityProvider().
		URL(gitlab.Spec.URL).
		ClientSecret(clientSecret).
		ClientID(gitlab.Spec.ClientID)

	if ca != "" {
		gitlabIDP.CA(ca)
	}

	return builder.Gitlab(gitlabIDP)
}

func init() {
	SchemeBuilder.Register(&GitLabIdentityProvider{}, &GitLabIdentityProviderList{})
}
