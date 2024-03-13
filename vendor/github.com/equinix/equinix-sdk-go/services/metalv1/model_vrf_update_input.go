/*
Metal API

# Introduction Equinix Metal provides a RESTful HTTP API which can be reached at <https://api.equinix.com/metal/v1>. This document describes the API and how to use it.  The API allows you to programmatically interact with all of your Equinix Metal resources, including devices, networks, addresses, organizations, projects, and your user account. Every feature of the Equinix Metal web interface is accessible through the API.  The API docs are generated from the Equinix Metal OpenAPI specification and are officially hosted at <https://metal.equinix.com/developers/api>.  # Common Parameters  The Equinix Metal API uses a few methods to minimize network traffic and improve throughput. These parameters are not used in all API calls, but are used often enough to warrant their own section. Look for these parameters in the documentation for the API calls that support them.  ## Pagination  Pagination is used to limit the number of results returned in a single request. The API will return a maximum of 100 results per page. To retrieve additional results, you can use the `page` and `per_page` query parameters.  The `page` parameter is used to specify the page number. The first page is `1`. The `per_page` parameter is used to specify the number of results per page. The maximum number of results differs by resource type.  ## Sorting  Where offered, the API allows you to sort results by a specific field. To sort results use the `sort_by` query parameter with the root level field name as the value. The `sort_direction` parameter is used to specify the sort direction, either either `asc` (ascending) or `desc` (descending).  ## Filtering  Filtering is used to limit the results returned in a single request. The API supports filtering by certain fields in the response. To filter results, you can use the field as a query parameter.  For example, to filter the IP list to only return public IPv4 addresses, you can filter by the `type` field, as in the following request:  ```sh curl -H 'X-Auth-Token: my_authentication_token' \\   https://api.equinix.com/metal/v1/projects/id/ips?type=public_ipv4 ```  Only IP addresses with the `type` field set to `public_ipv4` will be returned.  ## Searching  Searching is used to find matching resources using multiple field comparissons. The API supports searching in resources that define this behavior. Currently the search parameter is only available on devices, ssh_keys, api_keys and memberships endpoints.  To search resources you can use the `search` query parameter.  ## Include and Exclude  For resources that contain references to other resources, sucha as a Device that refers to the Project it resides in, the Equinix Metal API will returns `href` values (API links) to the associated resource.  ```json {   ...   \"project\": {     \"href\": \"/metal/v1/projects/f3f131c8-f302-49ef-8c44-9405022dc6dd\"   } } ```  If you're going need the project details, you can avoid a second API request.  Specify the contained `href` resources and collections that you'd like to have included in the response using the `include` query parameter.  For example:  ```sh curl -H 'X-Auth-Token: my_authentication_token' \\   https://api.equinix.com/metal/v1/user?include=projects ```  The `include` parameter is generally accepted in `GET`, `POST`, `PUT`, and `PATCH` requests where `href` resources are presented.  To have multiple resources include, use a comma-separated list (e.g. `?include=emails,projects,memberships`).  ```sh curl -H 'X-Auth-Token: my_authentication_token' \\   https://api.equinix.com/metal/v1/user?include=emails,projects,memberships ```  You may also include nested associations up to three levels deep using dot notation (`?include=memberships.projects`):  ```sh curl -H 'X-Auth-Token: my_authentication_token' \\   https://api.equinix.com/metal/v1/user?include=memberships.projects ```  To exclude resources, and optimize response delivery, use the `exclude` query parameter. The `exclude` parameter is generally accepted in `GET`, `POST`, `PUT`, and `PATCH` requests for fields with nested object responses. When excluded, these fields will be replaced with an object that contains only an `href` field.

API version: 1.0.0
Contact: support@equinixmetal.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package metalv1

import (
	"encoding/json"
)

// checks if the VrfUpdateInput type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &VrfUpdateInput{}

// VrfUpdateInput struct for VrfUpdateInput
type VrfUpdateInput struct {
	// Toggle to enable the dynamic bgp neighbors feature on the VRF
	BgpDynamicNeighborsEnabled *bool `json:"bgp_dynamic_neighbors_enabled,omitempty"`
	// Toggle to export the VRF route-map to the dynamic bgp neighbors
	BgpDynamicNeighborsExportRouteMap *bool `json:"bgp_dynamic_neighbors_export_route_map,omitempty"`
	// Toggle BFD on dynamic bgp neighbors sessions
	BgpDynamicNeighborsBfdEnabled *bool   `json:"bgp_dynamic_neighbors_bfd_enabled,omitempty"`
	Description                   *string `json:"description,omitempty"`
	// A list of CIDR network addresses. Like [\"10.0.0.0/16\", \"2001:d78::/56\"]. IPv4 blocks must be between /8 and /29 in size. IPv6 blocks must be between /56 and /64. A VRF\\'s IP ranges must be defined in order to create VRF IP Reservations, which can then be used for Metal Gateways or Virtual Circuits. Adding a new CIDR address to the list will result in the creation of a new IP Range for this VRF. Removal of an existing CIDR address from the list will result in the deletion of an existing IP Range for this VRF. Deleting an IP Range will result in the deletion of any VRF IP Reservations contained within the IP Range, as well as the VRF IP Reservation\\'s associated Metal Gateways or Virtual Circuits. If you do not wish to add or remove IP Ranges, either include the full existing list of IP Ranges in the update request, or do not specify the `ip_ranges` field in the update request. Specifying a value of `[]` will remove all existing IP Ranges from the VRF.
	IpRanges []string `json:"ip_ranges,omitempty"`
	// The new `local_asn` value for the VRF. This field cannot be updated when there are active Interconnection Virtual Circuits associated to the VRF, or if any of the VLANs of the VRF's metal gateway has been assigned on an instance.
	LocalAsn             *int32   `json:"local_asn,omitempty"`
	Name                 *string  `json:"name,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _VrfUpdateInput VrfUpdateInput

// NewVrfUpdateInput instantiates a new VrfUpdateInput object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewVrfUpdateInput() *VrfUpdateInput {
	this := VrfUpdateInput{}
	return &this
}

// NewVrfUpdateInputWithDefaults instantiates a new VrfUpdateInput object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewVrfUpdateInputWithDefaults() *VrfUpdateInput {
	this := VrfUpdateInput{}
	return &this
}

// GetBgpDynamicNeighborsEnabled returns the BgpDynamicNeighborsEnabled field value if set, zero value otherwise.
func (o *VrfUpdateInput) GetBgpDynamicNeighborsEnabled() bool {
	if o == nil || IsNil(o.BgpDynamicNeighborsEnabled) {
		var ret bool
		return ret
	}
	return *o.BgpDynamicNeighborsEnabled
}

// GetBgpDynamicNeighborsEnabledOk returns a tuple with the BgpDynamicNeighborsEnabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VrfUpdateInput) GetBgpDynamicNeighborsEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.BgpDynamicNeighborsEnabled) {
		return nil, false
	}
	return o.BgpDynamicNeighborsEnabled, true
}

// HasBgpDynamicNeighborsEnabled returns a boolean if a field has been set.
func (o *VrfUpdateInput) HasBgpDynamicNeighborsEnabled() bool {
	if o != nil && !IsNil(o.BgpDynamicNeighborsEnabled) {
		return true
	}

	return false
}

// SetBgpDynamicNeighborsEnabled gets a reference to the given bool and assigns it to the BgpDynamicNeighborsEnabled field.
func (o *VrfUpdateInput) SetBgpDynamicNeighborsEnabled(v bool) {
	o.BgpDynamicNeighborsEnabled = &v
}

// GetBgpDynamicNeighborsExportRouteMap returns the BgpDynamicNeighborsExportRouteMap field value if set, zero value otherwise.
func (o *VrfUpdateInput) GetBgpDynamicNeighborsExportRouteMap() bool {
	if o == nil || IsNil(o.BgpDynamicNeighborsExportRouteMap) {
		var ret bool
		return ret
	}
	return *o.BgpDynamicNeighborsExportRouteMap
}

// GetBgpDynamicNeighborsExportRouteMapOk returns a tuple with the BgpDynamicNeighborsExportRouteMap field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VrfUpdateInput) GetBgpDynamicNeighborsExportRouteMapOk() (*bool, bool) {
	if o == nil || IsNil(o.BgpDynamicNeighborsExportRouteMap) {
		return nil, false
	}
	return o.BgpDynamicNeighborsExportRouteMap, true
}

// HasBgpDynamicNeighborsExportRouteMap returns a boolean if a field has been set.
func (o *VrfUpdateInput) HasBgpDynamicNeighborsExportRouteMap() bool {
	if o != nil && !IsNil(o.BgpDynamicNeighborsExportRouteMap) {
		return true
	}

	return false
}

// SetBgpDynamicNeighborsExportRouteMap gets a reference to the given bool and assigns it to the BgpDynamicNeighborsExportRouteMap field.
func (o *VrfUpdateInput) SetBgpDynamicNeighborsExportRouteMap(v bool) {
	o.BgpDynamicNeighborsExportRouteMap = &v
}

// GetBgpDynamicNeighborsBfdEnabled returns the BgpDynamicNeighborsBfdEnabled field value if set, zero value otherwise.
func (o *VrfUpdateInput) GetBgpDynamicNeighborsBfdEnabled() bool {
	if o == nil || IsNil(o.BgpDynamicNeighborsBfdEnabled) {
		var ret bool
		return ret
	}
	return *o.BgpDynamicNeighborsBfdEnabled
}

// GetBgpDynamicNeighborsBfdEnabledOk returns a tuple with the BgpDynamicNeighborsBfdEnabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VrfUpdateInput) GetBgpDynamicNeighborsBfdEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.BgpDynamicNeighborsBfdEnabled) {
		return nil, false
	}
	return o.BgpDynamicNeighborsBfdEnabled, true
}

// HasBgpDynamicNeighborsBfdEnabled returns a boolean if a field has been set.
func (o *VrfUpdateInput) HasBgpDynamicNeighborsBfdEnabled() bool {
	if o != nil && !IsNil(o.BgpDynamicNeighborsBfdEnabled) {
		return true
	}

	return false
}

// SetBgpDynamicNeighborsBfdEnabled gets a reference to the given bool and assigns it to the BgpDynamicNeighborsBfdEnabled field.
func (o *VrfUpdateInput) SetBgpDynamicNeighborsBfdEnabled(v bool) {
	o.BgpDynamicNeighborsBfdEnabled = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *VrfUpdateInput) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VrfUpdateInput) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *VrfUpdateInput) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *VrfUpdateInput) SetDescription(v string) {
	o.Description = &v
}

// GetIpRanges returns the IpRanges field value if set, zero value otherwise.
func (o *VrfUpdateInput) GetIpRanges() []string {
	if o == nil || IsNil(o.IpRanges) {
		var ret []string
		return ret
	}
	return o.IpRanges
}

// GetIpRangesOk returns a tuple with the IpRanges field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VrfUpdateInput) GetIpRangesOk() ([]string, bool) {
	if o == nil || IsNil(o.IpRanges) {
		return nil, false
	}
	return o.IpRanges, true
}

// HasIpRanges returns a boolean if a field has been set.
func (o *VrfUpdateInput) HasIpRanges() bool {
	if o != nil && !IsNil(o.IpRanges) {
		return true
	}

	return false
}

// SetIpRanges gets a reference to the given []string and assigns it to the IpRanges field.
func (o *VrfUpdateInput) SetIpRanges(v []string) {
	o.IpRanges = v
}

// GetLocalAsn returns the LocalAsn field value if set, zero value otherwise.
func (o *VrfUpdateInput) GetLocalAsn() int32 {
	if o == nil || IsNil(o.LocalAsn) {
		var ret int32
		return ret
	}
	return *o.LocalAsn
}

// GetLocalAsnOk returns a tuple with the LocalAsn field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VrfUpdateInput) GetLocalAsnOk() (*int32, bool) {
	if o == nil || IsNil(o.LocalAsn) {
		return nil, false
	}
	return o.LocalAsn, true
}

// HasLocalAsn returns a boolean if a field has been set.
func (o *VrfUpdateInput) HasLocalAsn() bool {
	if o != nil && !IsNil(o.LocalAsn) {
		return true
	}

	return false
}

// SetLocalAsn gets a reference to the given int32 and assigns it to the LocalAsn field.
func (o *VrfUpdateInput) SetLocalAsn(v int32) {
	o.LocalAsn = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *VrfUpdateInput) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VrfUpdateInput) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *VrfUpdateInput) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *VrfUpdateInput) SetName(v string) {
	o.Name = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *VrfUpdateInput) GetTags() []string {
	if o == nil || IsNil(o.Tags) {
		var ret []string
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VrfUpdateInput) GetTagsOk() ([]string, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *VrfUpdateInput) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []string and assigns it to the Tags field.
func (o *VrfUpdateInput) SetTags(v []string) {
	o.Tags = v
}

func (o VrfUpdateInput) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o VrfUpdateInput) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.BgpDynamicNeighborsEnabled) {
		toSerialize["bgp_dynamic_neighbors_enabled"] = o.BgpDynamicNeighborsEnabled
	}
	if !IsNil(o.BgpDynamicNeighborsExportRouteMap) {
		toSerialize["bgp_dynamic_neighbors_export_route_map"] = o.BgpDynamicNeighborsExportRouteMap
	}
	if !IsNil(o.BgpDynamicNeighborsBfdEnabled) {
		toSerialize["bgp_dynamic_neighbors_bfd_enabled"] = o.BgpDynamicNeighborsBfdEnabled
	}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.IpRanges) {
		toSerialize["ip_ranges"] = o.IpRanges
	}
	if !IsNil(o.LocalAsn) {
		toSerialize["local_asn"] = o.LocalAsn
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *VrfUpdateInput) UnmarshalJSON(data []byte) (err error) {
	varVrfUpdateInput := _VrfUpdateInput{}

	err = json.Unmarshal(data, &varVrfUpdateInput)

	if err != nil {
		return err
	}

	*o = VrfUpdateInput(varVrfUpdateInput)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "bgp_dynamic_neighbors_enabled")
		delete(additionalProperties, "bgp_dynamic_neighbors_export_route_map")
		delete(additionalProperties, "bgp_dynamic_neighbors_bfd_enabled")
		delete(additionalProperties, "description")
		delete(additionalProperties, "ip_ranges")
		delete(additionalProperties, "local_asn")
		delete(additionalProperties, "name")
		delete(additionalProperties, "tags")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableVrfUpdateInput struct {
	value *VrfUpdateInput
	isSet bool
}

func (v NullableVrfUpdateInput) Get() *VrfUpdateInput {
	return v.value
}

func (v *NullableVrfUpdateInput) Set(val *VrfUpdateInput) {
	v.value = val
	v.isSet = true
}

func (v NullableVrfUpdateInput) IsSet() bool {
	return v.isSet
}

func (v *NullableVrfUpdateInput) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableVrfUpdateInput(val *VrfUpdateInput) *NullableVrfUpdateInput {
	return &NullableVrfUpdateInput{value: val, isSet: true}
}

func (v NullableVrfUpdateInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableVrfUpdateInput) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}