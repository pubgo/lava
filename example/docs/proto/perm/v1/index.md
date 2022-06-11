# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [perm/v1/group.proto](#perm_v1_group-proto)
    - [AddResForGroupRequest](#perm-v1-AddResForGroupRequest)
    - [AddResForGroupResponse](#perm-v1-AddResForGroupResponse)
    - [CreateGroupRequest](#perm-v1-CreateGroupRequest)
    - [CreateGroupResponse](#perm-v1-CreateGroupResponse)
    - [DelResForGroupRequest](#perm-v1-DelResForGroupRequest)
    - [DelResForGroupResponse](#perm-v1-DelResForGroupResponse)
    - [DeleteGroupRequest](#perm-v1-DeleteGroupRequest)
    - [DeleteGroupResponse](#perm-v1-DeleteGroupResponse)
    - [Group](#perm-v1-Group)
    - [ListGroupsRequest](#perm-v1-ListGroupsRequest)
    - [ListGroupsResponse](#perm-v1-ListGroupsResponse)
    - [MoveGroupRequest](#perm-v1-MoveGroupRequest)
    - [MoveGroupResponse](#perm-v1-MoveGroupResponse)
  
    - [GroupService](#perm-v1-GroupService)
  
- [perm/v1/menu.proto](#perm_v1_menu-proto)
    - [ListMenusRequest](#perm-v1-ListMenusRequest)
    - [ListMenusResponse](#perm-v1-ListMenusResponse)
    - [MenuItem](#perm-v1-MenuItem)
  
    - [MenuService](#perm-v1-MenuService)
  
- [perm/v1/org.proto](#perm_v1_org-proto)
    - [CreateOrgRequest](#perm-v1-CreateOrgRequest)
    - [CreateOrgResponse](#perm-v1-CreateOrgResponse)
    - [DeleteOrgRequest](#perm-v1-DeleteOrgRequest)
    - [DeleteOrgResponse](#perm-v1-DeleteOrgResponse)
    - [ListOrgRequest](#perm-v1-ListOrgRequest)
    - [ListOrgResponse](#perm-v1-ListOrgResponse)
    - [TransferOrgRequest](#perm-v1-TransferOrgRequest)
    - [TransferOrgResponse](#perm-v1-TransferOrgResponse)
  
    - [OrgService](#perm-v1-OrgService)
  
- [perm/v1/perm.proto](#perm_v1_perm-proto)
    - [EnforceRequest](#perm-v1-EnforceRequest)
    - [EnforceResponse](#perm-v1-EnforceResponse)
    - [PermGroup](#perm-v1-PermGroup)
    - [PermServiceListGroupsRequest](#perm-v1-PermServiceListGroupsRequest)
    - [PermServiceListGroupsResponse](#perm-v1-PermServiceListGroupsResponse)
    - [PermServiceListMenusRequest](#perm-v1-PermServiceListMenusRequest)
    - [PermServiceListMenusResponse](#perm-v1-PermServiceListMenusResponse)
    - [PermServiceListResourcesRequest](#perm-v1-PermServiceListResourcesRequest)
    - [PermServiceListResourcesResponse](#perm-v1-PermServiceListResourcesResponse)
    - [PermServiceSaveRolePermRequest](#perm-v1-PermServiceSaveRolePermRequest)
    - [PermServiceSaveRolePermRequest.Group](#perm-v1-PermServiceSaveRolePermRequest-Group)
    - [PermServiceSaveRolePermResponse](#perm-v1-PermServiceSaveRolePermResponse)
    - [Resource](#perm-v1-Resource)
  
    - [PermService](#perm-v1-PermService)
  
- [perm/v1/role.proto](#perm_v1_role-proto)
    - [AddRoleForUserRequest](#perm-v1-AddRoleForUserRequest)
    - [AddRoleForUserResponse](#perm-v1-AddRoleForUserResponse)
    - [CreateRoleRequest](#perm-v1-CreateRoleRequest)
    - [CreateRoleResponse](#perm-v1-CreateRoleResponse)
    - [DelRoleForUserRequest](#perm-v1-DelRoleForUserRequest)
    - [DelRoleForUserResponse](#perm-v1-DelRoleForUserResponse)
    - [DeleteRoleRequest](#perm-v1-DeleteRoleRequest)
    - [DeleteRoleResponse](#perm-v1-DeleteRoleResponse)
    - [GetRoleRequest](#perm-v1-GetRoleRequest)
    - [GetRoleResponse](#perm-v1-GetRoleResponse)
    - [GetRolesForUserRequest](#perm-v1-GetRolesForUserRequest)
    - [GetRolesForUserResponse](#perm-v1-GetRolesForUserResponse)
    - [GetUsersForRoleRequest](#perm-v1-GetUsersForRoleRequest)
    - [GetUsersForRoleResponse](#perm-v1-GetUsersForRoleResponse)
    - [ListRolesRequest](#perm-v1-ListRolesRequest)
    - [ListRolesResponse](#perm-v1-ListRolesResponse)
    - [Role](#perm-v1-Role)
    - [UpdateRoleRequest](#perm-v1-UpdateRoleRequest)
    - [UpdateRoleResponse](#perm-v1-UpdateRoleResponse)
    - [UserRoleRequest](#perm-v1-UserRoleRequest)
  
    - [RoleService](#perm-v1-RoleService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="perm_v1_group-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/v1/group.proto



<a name="perm-v1-AddResForGroupRequest"></a>

### AddResForGroupRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| group_type | [string](#string) |  |  |
| group_id | [string](#string) |  |  |
| res_id | [string](#string) |  |  |






<a name="perm-v1-AddResForGroupResponse"></a>

### AddResForGroupResponse







<a name="perm-v1-CreateGroupRequest"></a>

### CreateGroupRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| group_type | [string](#string) |  |  |
| group_id | [string](#string) |  |  |
| parent_group_type | [string](#string) |  |  |
| parent_group_id | [string](#string) |  |  |
| children | [string](#string) | repeated |  |






<a name="perm-v1-CreateGroupResponse"></a>

### CreateGroupResponse







<a name="perm-v1-DelResForGroupRequest"></a>

### DelResForGroupRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| group_type | [string](#string) |  |  |
| group_id | [string](#string) |  |  |
| res_id | [string](#string) |  |  |






<a name="perm-v1-DelResForGroupResponse"></a>

### DelResForGroupResponse







<a name="perm-v1-DeleteGroupRequest"></a>

### DeleteGroupRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| group_type | [string](#string) |  |  |
| group_id | [string](#string) |  |  |
| parent_group_type | [string](#string) |  |  |
| parent_group_id | [string](#string) |  |  |






<a name="perm-v1-DeleteGroupResponse"></a>

### DeleteGroupResponse







<a name="perm-v1-Group"></a>

### Group



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| res_type | [string](#string) |  |  |
| group_type | [string](#string) |  |  |
| group_id | [string](#string) |  |  |
| resources | [string](#string) | repeated |  |
| children | [Group](#perm-v1-Group) | repeated |  |






<a name="perm-v1-ListGroupsRequest"></a>

### ListGroupsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |






<a name="perm-v1-ListGroupsResponse"></a>

### ListGroupsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groups | [Group](#perm-v1-Group) | repeated |  |






<a name="perm-v1-MoveGroupRequest"></a>

### MoveGroupRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| from_group_type | [string](#string) |  |  |
| from_group_id | [string](#string) |  |  |
| cur_group_type | [string](#string) |  |  |
| cur_group_id | [string](#string) |  |  |
| to_group_type | [string](#string) |  |  |
| to_group_id | [string](#string) |  |  |






<a name="perm-v1-MoveGroupResponse"></a>

### MoveGroupResponse






 

 

 


<a name="perm-v1-GroupService"></a>

### GroupService
group grpc service, group is a tree managed by casbin

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateGroup | [CreateGroupRequest](#perm-v1-CreateGroupRequest) | [CreateGroupResponse](#perm-v1-CreateGroupResponse) | create group node, if parent is null, parent is root node each resource has a root under each org root node =&gt; group/{res_type}/org/org_{org_id} |
| DeleteGroup | [DeleteGroupRequest](#perm-v1-DeleteGroupRequest) | [DeleteGroupResponse](#perm-v1-DeleteGroupResponse) | delete group node, if parent is null, parent is root node |
| MoveGroup | [MoveGroupRequest](#perm-v1-MoveGroupRequest) | [MoveGroupResponse](#perm-v1-MoveGroupResponse) | move group node, if to_group is null, to_group is root node |
| ListGroups | [ListGroupsRequest](#perm-v1-ListGroupsRequest) | [ListGroupsResponse](#perm-v1-ListGroupsResponse) | list group tree The returned result is a tree structure {org_id} and {res_type} are a required |
| AddResForGroup | [AddResForGroupRequest](#perm-v1-AddResForGroupRequest) | [AddResForGroupResponse](#perm-v1-AddResForGroupResponse) | add resource to group node, if parent is null, parent is root node |
| DelResForGroup | [DelResForGroupRequest](#perm-v1-DelResForGroupRequest) | [DelResForGroupResponse](#perm-v1-DelResForGroupResponse) | delete resource from group node, if parent is null, parent is root node |

 



<a name="perm_v1_menu-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/v1/menu.proto



<a name="perm-v1-ListMenusRequest"></a>

### ListMenusRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| platform | [string](#string) |  |  |






<a name="perm-v1-ListMenusResponse"></a>

### ListMenusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| items | [MenuItem](#perm-v1-MenuItem) | repeated |  |






<a name="perm-v1-MenuItem"></a>

### MenuItem



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  |  |
| type | [string](#string) |  |  |
| name | [string](#string) |  |  |
| children | [MenuItem](#perm-v1-MenuItem) | repeated |  |





 

 

 


<a name="perm-v1-MenuService"></a>

### MenuService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListMenus | [ListMenusRequest](#perm-v1-ListMenusRequest) | [ListMenusResponse](#perm-v1-ListMenusResponse) | ListMenus returns available menus with hierarchy. |

 



<a name="perm_v1_org-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/v1/org.proto



<a name="perm-v1-CreateOrgRequest"></a>

### CreateOrgRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |






<a name="perm-v1-CreateOrgResponse"></a>

### CreateOrgResponse







<a name="perm-v1-DeleteOrgRequest"></a>

### DeleteOrgRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |






<a name="perm-v1-DeleteOrgResponse"></a>

### DeleteOrgResponse







<a name="perm-v1-ListOrgRequest"></a>

### ListOrgRequest







<a name="perm-v1-ListOrgResponse"></a>

### ListOrgResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| orgs | [string](#string) | repeated |  |






<a name="perm-v1-TransferOrgRequest"></a>

### TransferOrgRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| new_user_id | [string](#string) |  |  |






<a name="perm-v1-TransferOrgResponse"></a>

### TransferOrgResponse






 

 

 


<a name="perm-v1-OrgService"></a>

### OrgService
org grpc service, OrgSrv provides org info inside RBAC instead of global

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateOrg | [CreateOrgRequest](#perm-v1-CreateOrgRequest) | [CreateOrgResponse](#perm-v1-CreateOrgResponse) | CreateOrg init org, create org role and bind all function permissions, the method is idempotent {org_id} is required when {user_id} is set, {user_id} will be admin |
| DeleteOrg | [DeleteOrgRequest](#perm-v1-DeleteOrgRequest) | [DeleteOrgResponse](#perm-v1-DeleteOrgResponse) | delete org all perms and data {org_id} is required |
| TransferOrg | [TransferOrgRequest](#perm-v1-TransferOrgRequest) | [TransferOrgResponse](#perm-v1-TransferOrgResponse) | transfer org admin to {new_user_id}, then {user_id} will only lose the admin role, and other roles will be retained all parameters are required |
| ListOrg | [ListOrgRequest](#perm-v1-ListOrgRequest) | [ListOrgResponse](#perm-v1-ListOrgResponse) | list all org |

 



<a name="perm_v1_perm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/v1/perm.proto



<a name="perm-v1-EnforceRequest"></a>

### EnforceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  | organization |
| user_id | [string](#string) |  | user_id or role_id must be selected |
| role_id | [string](#string) |  |  |
| res_type | [string](#string) |  | resource type, e.g. &#34;box&#34;, &#34;camera&#34;, &#34;api&#34;, etc. |
| group_type | [string](#string) |  |  |
| res_id | [string](#string) |  | resource id, e.g. &#34;123&#34;, &#34;/api/cameras&#34;, etc. |
| act | [string](#string) |  | resource act, e.g. &#34;post&#34;, &#34;export&#34;, &#34;play&#34;, &#34;download&#34;, etc. |






<a name="perm-v1-EnforceResponse"></a>

### EnforceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ok | [bool](#bool) |  |  |
| code | [string](#string) |  |  |






<a name="perm-v1-PermGroup"></a>

### PermGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| res_type | [string](#string) |  |  |
| group_type | [string](#string) |  |  |
| group_id | [string](#string) |  |  |
| resources | [string](#string) | repeated |  |
| children | [PermGroup](#perm-v1-PermGroup) | repeated |  |
| acts | [string](#string) | repeated |  |






<a name="perm-v1-PermServiceListGroupsRequest"></a>

### PermServiceListGroupsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| act | [string](#string) |  |  |






<a name="perm-v1-PermServiceListGroupsResponse"></a>

### PermServiceListGroupsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groups | [PermGroup](#perm-v1-PermGroup) | repeated |  |






<a name="perm-v1-PermServiceListMenusRequest"></a>

### PermServiceListMenusRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| platform | [string](#string) |  |  |






<a name="perm-v1-PermServiceListMenusResponse"></a>

### PermServiceListMenusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| items | [MenuItem](#perm-v1-MenuItem) | repeated |  |






<a name="perm-v1-PermServiceListResourcesRequest"></a>

### PermServiceListResourcesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| act | [string](#string) |  |  |






<a name="perm-v1-PermServiceListResourcesResponse"></a>

### PermServiceListResourcesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resources | [Resource](#perm-v1-Resource) | repeated |  |






<a name="perm-v1-PermServiceSaveRolePermRequest"></a>

### PermServiceSaveRolePermRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| menus | [string](#string) | repeated |  |
| groups | [PermServiceSaveRolePermRequest.Group](#perm-v1-PermServiceSaveRolePermRequest-Group) | repeated |  |






<a name="perm-v1-PermServiceSaveRolePermRequest-Group"></a>

### PermServiceSaveRolePermRequest.Group



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| res_type | [string](#string) |  |  |
| group_type | [string](#string) |  |  |
| group_id | [string](#string) |  |  |






<a name="perm-v1-PermServiceSaveRolePermResponse"></a>

### PermServiceSaveRolePermResponse







<a name="perm-v1-Resource"></a>

### Resource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| res_id | [string](#string) |  |  |
| acts | [string](#string) | repeated |  |





 

 

 


<a name="perm-v1-PermService"></a>

### PermService
Perm grpc service, Perm for authentication and resource list

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Enforce | [EnforceRequest](#perm-v1-EnforceRequest) | [EnforceResponse](#perm-v1-EnforceResponse) |  |
| ListResources | [PermServiceListResourcesRequest](#perm-v1-PermServiceListResourcesRequest) | [PermServiceListResourcesResponse](#perm-v1-PermServiceListResourcesResponse) |  |
| ListMenus | [PermServiceListMenusRequest](#perm-v1-PermServiceListMenusRequest) | [PermServiceListMenusResponse](#perm-v1-PermServiceListMenusResponse) |  |
| ListGroups | [PermServiceListGroupsRequest](#perm-v1-PermServiceListGroupsRequest) | [PermServiceListGroupsResponse](#perm-v1-PermServiceListGroupsResponse) |  |
| SaveRolePerm | [PermServiceSaveRolePermRequest](#perm-v1-PermServiceSaveRolePermRequest) | [PermServiceSaveRolePermResponse](#perm-v1-PermServiceSaveRolePermResponse) |  |

 



<a name="perm_v1_role-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/v1/role.proto



<a name="perm-v1-AddRoleForUserRequest"></a>

### AddRoleForUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |






<a name="perm-v1-AddRoleForUserResponse"></a>

### AddRoleForUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ok | [bool](#bool) |  |  |






<a name="perm-v1-CreateRoleRequest"></a>

### CreateRoleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-CreateRoleResponse"></a>

### CreateRoleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-DelRoleForUserRequest"></a>

### DelRoleForUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |






<a name="perm-v1-DelRoleForUserResponse"></a>

### DelRoleForUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ok | [bool](#bool) |  |  |






<a name="perm-v1-DeleteRoleRequest"></a>

### DeleteRoleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| name | [string](#string) |  |  |
| org_id | [string](#string) |  |  |






<a name="perm-v1-DeleteRoleResponse"></a>

### DeleteRoleResponse







<a name="perm-v1-GetRoleRequest"></a>

### GetRoleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |






<a name="perm-v1-GetRoleResponse"></a>

### GetRoleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-GetRolesForUserRequest"></a>

### GetRolesForUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |






<a name="perm-v1-GetRolesForUserResponse"></a>

### GetRolesForUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| roles | [string](#string) | repeated |  |






<a name="perm-v1-GetUsersForRoleRequest"></a>

### GetUsersForRoleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |






<a name="perm-v1-GetUsersForRoleResponse"></a>

### GetUsersForRoleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [string](#string) | repeated |  |






<a name="perm-v1-ListRolesRequest"></a>

### ListRolesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |






<a name="perm-v1-ListRolesResponse"></a>

### ListRolesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| roles | [Role](#perm-v1-Role) | repeated |  |






<a name="perm-v1-Role"></a>

### Role



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  | role id |
| name | [string](#string) |  | role name, e.g. &#34;admin or 123456&#34; |
| status | [string](#string) |  | role status |
| org_id | [string](#string) |  | org id, |
| display_name | [string](#string) |  | role display name, e.g. &#34;administrators&#34; |






<a name="perm-v1-UpdateRoleRequest"></a>

### UpdateRoleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-UpdateRoleResponse"></a>

### UpdateRoleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-UserRoleRequest"></a>

### UserRoleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |





 

 

 


<a name="perm-v1-RoleService"></a>

### RoleService
role grpc service, RoleService provides role management and user role management

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateRole | [CreateRoleRequest](#perm-v1-CreateRoleRequest) | [CreateRoleResponse](#perm-v1-CreateRoleResponse) | create role |
| DeleteRole | [DeleteRoleRequest](#perm-v1-DeleteRoleRequest) | [DeleteRoleResponse](#perm-v1-DeleteRoleResponse) | - delete role by id or name - req: id=12 - req: name=admin,org_id=ka |
| UpdateRole | [UpdateRoleRequest](#perm-v1-UpdateRoleRequest) | [UpdateRoleResponse](#perm-v1-UpdateRoleResponse) | - update role by id or name - req: id=12 - req: name=admin,org_id=ka |
| GetRole | [GetRoleRequest](#perm-v1-GetRoleRequest) | [GetRoleResponse](#perm-v1-GetRoleResponse) | get role by id |
| ListRoles | [ListRolesRequest](#perm-v1-ListRolesRequest) | [ListRolesResponse](#perm-v1-ListRolesResponse) | list role by org_id |
| AddRoleForUser | [AddRoleForUserRequest](#perm-v1-AddRoleForUserRequest) | [AddRoleForUserResponse](#perm-v1-AddRoleForUserResponse) | add role to user all parameters are required |
| DelRoleForUser | [DelRoleForUserRequest](#perm-v1-DelRoleForUserRequest) | [DelRoleForUserResponse](#perm-v1-DelRoleForUserResponse) | delete user org all parameters are required if {role_id} is *, it will delete all role about the user |
| GetRolesForUser | [GetRolesForUserRequest](#perm-v1-GetRolesForUserRequest) | [GetRolesForUserResponse](#perm-v1-GetRolesForUserResponse) | get user all roles {org_id} and {user_id} are required |
| GetUsersForRole | [GetUsersForRoleRequest](#perm-v1-GetUsersForRoleRequest) | [GetUsersForRoleResponse](#perm-v1-GetUsersForRoleResponse) | get users from a {role_id} if {role_id} is null, you will get all users of the {org_id} |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

