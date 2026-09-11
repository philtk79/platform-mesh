/*
Copyright The Platform Mesh Authors.

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
	"go.platform-mesh.io/subroutines/conditions"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IdentityProviderClientType string

const (
	IdentityProviderClientTypeConfidential IdentityProviderClientType = "confidential"
	IdentityProviderClientTypePublic       IdentityProviderClientType = "public"
)

type IdentityProviderClientConfig struct {
	// +kubebuilder:validation:Enum=confidential;public
	ClientType             IdentityProviderClientType `json:"clientType"`
	ClientName             string                     `json:"clientName"`
	RedirectURIs           []string                   `json:"redirectUris"`
	PostLogoutRedirectURIs []string                   `json:"postLogoutRedirectUris,omitempty"`
	SecretRef              corev1.SecretReference     `json:"secretRef,omitempty"`
}

type UpstreamIdentityProviderType string

const (
	UpstreamIdentityProviderTypeOIDC UpstreamIdentityProviderType = "oidc"
)

// UpstreamIdentityProvider holds fields independent from a provider's type
// and references provider-specific information under OIDC. Security-relevant
// broker settings (signature validation, sync mode, login flows, mappers) are
// operator-controlled and intentionally absent here.
type UpstreamIdentityProvider struct {
	Alias           string `json:"alias"`
	DisplayName     string `json:"displayName,omitempty"`
	Enabled         *bool  `json:"enabled,omitempty"`
	HideOnLoginPage *bool  `json:"hideOnLoginPage,omitempty"`
	// EmailDomainRouting configures email-domain based identity-first login for
	// this upstream IdP. It is provider-agnostic; the operator translates it to
	// the backend's routing mechanism (for Keycloak, Organizations).
	EmailDomainRouting *EmailDomainRouting `json:"emailDomainRouting,omitempty"`
	// +kubebuilder:validation:Enum=oidc
	Type UpstreamIdentityProviderType `json:"type"`
	OIDC *OIDCUpstreamConfig          `json:"oidc,omitempty"`
}

// EmailDomainRouting configures email-domain based identity-first login for an
// upstream identity provider: users are routed to the provider based on their
// email domain. It is intentionally provider-agnostic — the security-operator
// maps it onto the concrete backend (for Keycloak: Organizations and the
// kc.org.* broker settings).
//
// Warning: removing all domains or deleting the parent resource clears the broker's
// organization linkage and deletes the linked Keycloak organization. That
// organization may be shared if multiple upstream IdPs reference the same domains.
//
// Domain ownership is not verified at admission time. A future release may require
// org owners to prove control of a domain before it can be linked.
type EmailDomainRouting struct {
	// Domains lists the email domains routed to this upstream IdP. Each entry is
	// normalized to lowercase. Org owners should only add domains their organization
	// controls; verification is not enforced yet.
	// +kubebuilder:validation:MinItems=1
	Domains []string `json:"domains"`
	// AutoRedirect immediately sends users whose email domain matches straight to
	// this upstream IdP instead of showing the provider-selection screen.
	AutoRedirect *bool `json:"autoRedirect,omitempty"`
	// HideUntilDomainMatch hides this upstream IdP on the login page until a
	// matching email domain routes a user to it.
	HideUntilDomainMatch *bool `json:"hideUntilDomainMatch,omitempty"`
}

// OIDCUpstreamConfig holds OIDC-specific upstream identity provider
// configuration. Signature validation and client authentication are pinned by
// the operator and cannot be configured here.
type OIDCUpstreamConfig struct {
	// Either DiscoveryURL or the manual endpoint fields need to be set.
	DiscoveryURL     string `json:"discoveryUrl,omitempty"`
	Issuer           string `json:"issuer,omitempty"`
	AuthorizationURL string `json:"authorizationUrl,omitempty"`
	TokenURL         string `json:"tokenUrl,omitempty"`
	JWKSURL          string `json:"jwksUrl,omitempty"`
	ClientID         string `json:"clientId,omitempty"`
}

// IdentityProviderConfigurationSpec defines the desired state of IdentityProviderConfiguration
type IdentityProviderConfigurationSpec struct {
	RegistrationAllowed bool                           `json:"registrationAllowed,omitempty"`
	Clients             []IdentityProviderClientConfig `json:"clients"`
}

// ManagedClient tracks a client that is managed by the operator.
type ManagedClient struct {
	ClientID              string                 `json:"clientId"`
	RegistrationClientURI string                 `json:"registrationClientUri"`
	SecretRef             corev1.SecretReference `json:"secretRef"`
}

// IdentityProviderConfigurationStatus defines the observed state of IdentityProviderConfiguration.
type IdentityProviderConfigurationStatus struct {
	Conditions     []metav1.Condition       `json:"conditions,omitempty"`
	ManagedClients map[string]ManagedClient `json:"managedClients,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// IdentityProviderConfiguration is the Schema for the identityproviderconfigurations API
type IdentityProviderConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of IdentityProviderConfiguration
	// +required
	Spec IdentityProviderConfigurationSpec `json:"spec"`

	// status defines the observed state of IdentityProviderConfiguration
	// +optional
	Status IdentityProviderConfigurationStatus `json:"status,omitempty,omitzero"`
}

// GetConditions implements conditions.ConditionAccessor.
func (in *IdentityProviderConfiguration) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions implements conditions.ConditionAccessor.
func (in *IdentityProviderConfiguration) SetConditions(c []metav1.Condition) {
	in.Status.Conditions = c
}

// +kubebuilder:object:root=true

// IdentityProviderConfigurationList contains a list of IdentityProviderConfiguration
type IdentityProviderConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityProviderConfiguration `json:"items"`
}

var _ conditions.ConditionAccessor = &IdentityProviderConfiguration{}
