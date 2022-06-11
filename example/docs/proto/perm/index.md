# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [perm/catalog.proto](#perm_catalog-proto)
    - [AddResForCatalogReq](#perm-v1-AddResForCatalogReq)
    - [AddResForCatalogResp](#perm-v1-AddResForCatalogResp)
    - [AddRoleCatalogReq](#perm-v1-AddRoleCatalogReq)
    - [AddRoleCatalogResp](#perm-v1-AddRoleCatalogResp)
    - [Catalog](#perm-v1-Catalog)
    - [Catalog.ActsEntry](#perm-v1-Catalog-ActsEntry)
    - [Catalog.ChildrenEntry](#perm-v1-Catalog-ChildrenEntry)
    - [Catalog.NodesEntry](#perm-v1-Catalog-NodesEntry)
    - [CreateCatalogReq](#perm-v1-CreateCatalogReq)
    - [CreateCatalogResp](#perm-v1-CreateCatalogResp)
    - [DelCatalogReq](#perm-v1-DelCatalogReq)
    - [DelCatalogResp](#perm-v1-DelCatalogResp)
    - [DelResForCatalogReq](#perm-v1-DelResForCatalogReq)
    - [DelResForCatalogResp](#perm-v1-DelResForCatalogResp)
    - [DelRoleCatalogReq](#perm-v1-DelRoleCatalogReq)
    - [DelRoleCatalogResp](#perm-v1-DelRoleCatalogResp)
    - [GetCatalogReq](#perm-v1-GetCatalogReq)
    - [GetCatalogResp](#perm-v1-GetCatalogResp)
    - [GetCatalogsByRoleReq](#perm-v1-GetCatalogsByRoleReq)
    - [GetRolesByCatalogResp](#perm-v1-GetRolesByCatalogResp)
    - [ListCatalogsReq](#perm-v1-ListCatalogsReq)
    - [ListCatalogsResp](#perm-v1-ListCatalogsResp)
    - [MoveCatalogReq](#perm-v1-MoveCatalogReq)
    - [MoveCatalogResp](#perm-v1-MoveCatalogResp)
    - [UpdateCatalogReq](#perm-v1-UpdateCatalogReq)
    - [UpdateCatalogResp](#perm-v1-UpdateCatalogResp)
  
    - [CatalogSrv](#perm-v1-CatalogSrv)
  
- [perm/common.proto](#perm_common-proto)
    - [BoolResp](#perm-v1-BoolResp)
    - [Empty](#perm-v1-Empty)
    - [List](#perm-v1-List)
  
    - [platform](#perm-v1-platform)
    - [resType](#perm-v1-resType)
    - [status](#perm-v1-status)
  
- [perm/debug.proto](#perm_debug-proto)
    - [GetUserInfoReq](#perm-v1-GetUserInfoReq)
    - [GetUserInfoResp](#perm-v1-GetUserInfoResp)
    - [GetUserInfoResp.CatalogsEntry](#perm-v1-GetUserInfoResp-CatalogsEntry)
    - [GetUserInfoResp.MethodRulesEntry](#perm-v1-GetUserInfoResp-MethodRulesEntry)
    - [GetUserInfoResp.ResourcesEntry](#perm-v1-GetUserInfoResp-ResourcesEntry)
    - [GetUserInfoResp.RolesEntry](#perm-v1-GetUserInfoResp-RolesEntry)
  
    - [DebugSrv](#perm-v1-DebugSrv)
  
- [perm/method-rule.proto](#perm_method-rule-proto)
    - [CreateMethodRuleReq](#perm-v1-CreateMethodRuleReq)
    - [CreateMethodRuleResp](#perm-v1-CreateMethodRuleResp)
    - [DelMethodRuleReq](#perm-v1-DelMethodRuleReq)
    - [DelMethodRuleResp](#perm-v1-DelMethodRuleResp)
    - [GetMethodRuleReq](#perm-v1-GetMethodRuleReq)
    - [GetMethodRuleResp](#perm-v1-GetMethodRuleResp)
    - [ListMethodRulesReq](#perm-v1-ListMethodRulesReq)
    - [ListMethodRulesResp](#perm-v1-ListMethodRulesResp)
    - [MenuTree](#perm-v1-MenuTree)
    - [MethodRule](#perm-v1-MethodRule)
    - [SaveMethodRuleReq](#perm-v1-SaveMethodRuleReq)
    - [UpdateMethodRuleReq](#perm-v1-UpdateMethodRuleReq)
    - [UpdateMethodRuleResp](#perm-v1-UpdateMethodRuleResp)
  
    - [MethodRuleSrv](#perm-v1-MethodRuleSrv)
  
- [perm/org.proto](#perm_org-proto)
    - [GetOrgResp](#perm-v1-GetOrgResp)
    - [GetOrgResp.PermsEntry](#perm-v1-GetOrgResp-PermsEntry)
    - [GetOrgResp.RolesEntry](#perm-v1-GetOrgResp-RolesEntry)
    - [ListOrgResp](#perm-v1-ListOrgResp)
    - [OrgReq](#perm-v1-OrgReq)
    - [TransferOrgReq](#perm-v1-TransferOrgReq)
  
    - [OrgSrv](#perm-v1-OrgSrv)
  
- [perm/perm.proto](#perm_perm-proto)
    - [EnforceReq](#perm-v1-EnforceReq)
    - [EnforceResp](#perm-v1-EnforceResp)
    - [ListResReq](#perm-v1-ListResReq)
    - [ListResResp](#perm-v1-ListResResp)
    - [ListResResp.ResourcesEntry](#perm-v1-ListResResp-ResourcesEntry)
    - [SaveRolePermReq](#perm-v1-SaveRolePermReq)
    - [SaveRolePermReq.Catalog](#perm-v1-SaveRolePermReq-Catalog)
  
    - [Perm](#perm-v1-Perm)
  
- [perm/role.proto](#perm_role-proto)
    - [CreateRoleReq](#perm-v1-CreateRoleReq)
    - [CreateRoleResp](#perm-v1-CreateRoleResp)
    - [DelRoleReq](#perm-v1-DelRoleReq)
    - [DelRoleResp](#perm-v1-DelRoleResp)
    - [GetRolePermReq](#perm-v1-GetRolePermReq)
    - [GetRolePermResp](#perm-v1-GetRolePermResp)
    - [GetRolePermResp.CatalogsEntry](#perm-v1-GetRolePermResp-CatalogsEntry)
    - [GetRoleReq](#perm-v1-GetRoleReq)
    - [GetRoleResp](#perm-v1-GetRoleResp)
    - [GetRolesForUserResp](#perm-v1-GetRolesForUserResp)
    - [GetUsersForRoleResp](#perm-v1-GetUsersForRoleResp)
    - [ListRolesReq](#perm-v1-ListRolesReq)
    - [ListRolesResp](#perm-v1-ListRolesResp)
    - [PermReq](#perm-v1-PermReq)
    - [Role](#perm-v1-Role)
    - [UpdateRoleReq](#perm-v1-UpdateRoleReq)
    - [UpdateRoleResp](#perm-v1-UpdateRoleResp)
    - [UserRoleReq](#perm-v1-UserRoleReq)
  
    - [RoleSrv](#perm-v1-RoleSrv)
  
- [Scalar Value Types](#scalar-value-types)



<a name="perm_catalog-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/catalog.proto



<a name="perm-v1-AddResForCatalogReq"></a>

### AddResForCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| catalog_id | [string](#string) |  |  |
| node_type | [string](#string) |  |  |
| res_id | [string](#string) |  |  |






<a name="perm-v1-AddResForCatalogResp"></a>

### AddResForCatalogResp







<a name="perm-v1-AddRoleCatalogReq"></a>

### AddRoleCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  |  |
| catalog_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |






<a name="perm-v1-AddRoleCatalogResp"></a>

### AddRoleCatalogResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| catalog_id | [string](#string) |  |  |






<a name="perm-v1-Catalog"></a>

### Catalog



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  |  |
| org_id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| node_type | [string](#string) |  |  |
| children | [Catalog.ChildrenEntry](#perm-v1-Catalog-ChildrenEntry) | repeated |  |
| nodes | [Catalog.NodesEntry](#perm-v1-Catalog-NodesEntry) | repeated |  |
| acts | [Catalog.ActsEntry](#perm-v1-Catalog-ActsEntry) | repeated |  |






<a name="perm-v1-Catalog-ActsEntry"></a>

### Catalog.ActsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="perm-v1-Catalog-ChildrenEntry"></a>

### Catalog.ChildrenEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="perm-v1-Catalog-NodesEntry"></a>

### Catalog.NodesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Catalog](#perm-v1-Catalog) |  |  |






<a name="perm-v1-CreateCatalogReq"></a>

### CreateCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| name | [string](#string) |  | name is res_id |
| node_type | [string](#string) |  |  |
| parent | [string](#string) |  | parent is parent res_id |
| parent_node_type | [string](#string) |  |  |
| children | [string](#string) | repeated |  |






<a name="perm-v1-CreateCatalogResp"></a>

### CreateCatalogResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [Catalog](#perm-v1-Catalog) |  |  |






<a name="perm-v1-DelCatalogReq"></a>

### DelCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| name | [string](#string) |  | name is res_id |
| node_type | [string](#string) |  |  |
| parent | [string](#string) |  | parent is parent res_id |
| parent_node_type | [string](#string) |  |  |






<a name="perm-v1-DelCatalogResp"></a>

### DelCatalogResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [Catalog](#perm-v1-Catalog) |  |  |






<a name="perm-v1-DelResForCatalogReq"></a>

### DelResForCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| catalog_id | [string](#string) |  |  |
| node_type | [string](#string) |  |  |
| res_id | [string](#string) |  |  |






<a name="perm-v1-DelResForCatalogResp"></a>

### DelResForCatalogResp







<a name="perm-v1-DelRoleCatalogReq"></a>

### DelRoleCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="perm-v1-DelRoleCatalogResp"></a>

### DelRoleCatalogResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| catalog_id | [string](#string) |  |  |






<a name="perm-v1-GetCatalogReq"></a>

### GetCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="perm-v1-GetCatalogResp"></a>

### GetCatalogResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [Catalog](#perm-v1-Catalog) |  |  |






<a name="perm-v1-GetCatalogsByRoleReq"></a>

### GetCatalogsByRoleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  |  |






<a name="perm-v1-GetRolesByCatalogResp"></a>

### GetRolesByCatalogResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog_id | [string](#string) |  |  |






<a name="perm-v1-ListCatalogsReq"></a>

### ListCatalogsReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| act | [string](#string) |  |  |






<a name="perm-v1-ListCatalogsResp"></a>

### ListCatalogsResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalogs | [Catalog](#perm-v1-Catalog) | repeated |  |






<a name="perm-v1-MoveCatalogReq"></a>

### MoveCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| from_catalog_id | [string](#string) |  |  |
| from_node_type | [string](#string) |  |  |
| cur_node_type | [string](#string) |  |  |
| cur_catalog_id | [string](#string) |  |  |
| to_catalog_id | [string](#string) |  |  |
| to_node_type | [string](#string) |  |  |






<a name="perm-v1-MoveCatalogResp"></a>

### MoveCatalogResp







<a name="perm-v1-UpdateCatalogReq"></a>

### UpdateCatalogReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [Catalog](#perm-v1-Catalog) |  |  |






<a name="perm-v1-UpdateCatalogResp"></a>

### UpdateCatalogResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [Catalog](#perm-v1-Catalog) |  |  |





 

 

 


<a name="perm-v1-CatalogSrv"></a>

### CatalogSrv
catalog grpc service, catalog is a tree managed by casbin

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateCatalog | [CreateCatalogReq](#perm-v1-CreateCatalogReq) | [BoolResp](#perm-v1-BoolResp) | create catalog node, if parent is null, parent is root node each resource has a root under each org root node =&gt; group/{res_type}/org/org_{org_id} |
| DelCatalog | [DelCatalogReq](#perm-v1-DelCatalogReq) | [BoolResp](#perm-v1-BoolResp) | delete catalog node, if parent is null, parent is root node |
| MoveCatalog | [MoveCatalogReq](#perm-v1-MoveCatalogReq) | [BoolResp](#perm-v1-BoolResp) | move catalog node, if to_catalog is null, to_catalog is root node |
| ListCatalogs | [ListCatalogsReq](#perm-v1-ListCatalogsReq) | [ListCatalogsResp](#perm-v1-ListCatalogsResp) | list catalog tree The returned result is a tree structure The {acts} in the return value list is null {org_id} and {res_type} are a required |
| AddResForCatalog | [AddResForCatalogReq](#perm-v1-AddResForCatalogReq) | [BoolResp](#perm-v1-BoolResp) | add resource to catalog node, if parent is null, parent is root node |
| DelResForCatalog | [DelResForCatalogReq](#perm-v1-DelResForCatalogReq) | [BoolResp](#perm-v1-BoolResp) | delete resource from catalog node, if parent is null, parent is root node |

 



<a name="perm_common-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/common.proto



<a name="perm-v1-BoolResp"></a>

### BoolResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ok | [bool](#bool) |  |  |






<a name="perm-v1-Empty"></a>

### Empty







<a name="perm-v1-List"></a>

### List



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [string](#string) | repeated |  |





 


<a name="perm-v1-platform"></a>

### platform


| Name | Number | Description |
| ---- | ------ | ----------- |
| pc | 0 |  |
| mobile | 1 |  |
| ka | 2 |  |
| vision | 3 |  |



<a name="perm-v1-resType"></a>

### resType


| Name | Number | Description |
| ---- | ------ | ----------- |
| api | 0 |  |
| action | 1 |  |
| menu | 2 |  |
| user | 3 |  |
| box | 4 |  |
| site | 5 |  |
| camera | 6 |  |



<a name="perm-v1-status"></a>

### status


| Name | Number | Description |
| ---- | ------ | ----------- |
| disabled | 0 |  |
| enabled | 1 |  |
| deleted | 2 |  |


 

 

 



<a name="perm_debug-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/debug.proto



<a name="perm-v1-GetUserInfoReq"></a>

### GetUserInfoReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |






<a name="perm-v1-GetUserInfoResp"></a>

### GetUserInfoResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| domains | [string](#string) | repeated |  |
| roles | [GetUserInfoResp.RolesEntry](#perm-v1-GetUserInfoResp-RolesEntry) | repeated |  |
| method_rules | [GetUserInfoResp.MethodRulesEntry](#perm-v1-GetUserInfoResp-MethodRulesEntry) | repeated |  |
| catalogs | [GetUserInfoResp.CatalogsEntry](#perm-v1-GetUserInfoResp-CatalogsEntry) | repeated |  |
| resources | [GetUserInfoResp.ResourcesEntry](#perm-v1-GetUserInfoResp-ResourcesEntry) | repeated |  |






<a name="perm-v1-GetUserInfoResp-CatalogsEntry"></a>

### GetUserInfoResp.CatalogsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ListCatalogsResp](#perm-v1-ListCatalogsResp) |  |  |






<a name="perm-v1-GetUserInfoResp-MethodRulesEntry"></a>

### GetUserInfoResp.MethodRulesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ListMethodRulesResp](#perm-v1-ListMethodRulesResp) |  |  |






<a name="perm-v1-GetUserInfoResp-ResourcesEntry"></a>

### GetUserInfoResp.ResourcesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ListResResp](#perm-v1-ListResResp) |  |  |






<a name="perm-v1-GetUserInfoResp-RolesEntry"></a>

### GetUserInfoResp.RolesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ListRolesResp](#perm-v1-ListRolesResp) |  |  |





 

 

 


<a name="perm-v1-DebugSrv"></a>

### DebugSrv
debug grpc service, DebugSrv just debug for casbin, No external services

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetUserInfo | [GetUserInfoReq](#perm-v1-GetUserInfoReq) | [GetUserInfoResp](#perm-v1-GetUserInfoResp) |  |

 



<a name="perm_method-rule-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/method-rule.proto



<a name="perm-v1-CreateMethodRuleReq"></a>

### CreateMethodRuleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rule | [MethodRule](#perm-v1-MethodRule) |  |  |






<a name="perm-v1-CreateMethodRuleResp"></a>

### CreateMethodRuleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rule | [MethodRule](#perm-v1-MethodRule) |  |  |






<a name="perm-v1-DelMethodRuleReq"></a>

### DelMethodRuleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [uint32](#uint32) |  |  |






<a name="perm-v1-DelMethodRuleResp"></a>

### DelMethodRuleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rule | [MethodRule](#perm-v1-MethodRule) |  |  |






<a name="perm-v1-GetMethodRuleReq"></a>

### GetMethodRuleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |






<a name="perm-v1-GetMethodRuleResp"></a>

### GetMethodRuleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rule | [MethodRule](#perm-v1-MethodRule) |  |  |






<a name="perm-v1-ListMethodRulesReq"></a>

### ListMethodRulesReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| platform | [string](#string) |  |  |
| org_id | [string](#string) |  |  |






<a name="perm-v1-ListMethodRulesResp"></a>

### ListMethodRulesResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rules | [MethodRule](#perm-v1-MethodRule) | repeated |  |






<a name="perm-v1-MenuTree"></a>

### MenuTree



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| code | [string](#string) |  |  |
| parent_code | [string](#string) |  |  |
| platform | [string](#string) |  |  |






<a name="perm-v1-MethodRule"></a>

### MethodRule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| method | [string](#string) |  | restful api method |
| path | [string](#string) |  | restful api path |
| res_type | [string](#string) |  |  |
| display_name | [string](#string) |  |  |
| code | [string](#string) |  |  |
| target_type | [string](#string) |  | The resource types involved in the API backend implementation |
| created_at | [int64](#int64) |  |  |
| updated_at | [int64](#int64) |  |  |
| deleted_at | [int64](#int64) |  |  |
| org_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| children | [MethodRule](#perm-v1-MethodRule) | repeated |  |
| parent_code | [string](#string) | repeated |  |






<a name="perm-v1-SaveMethodRuleReq"></a>

### SaveMethodRuleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rules | [string](#string) | repeated |  |






<a name="perm-v1-UpdateMethodRuleReq"></a>

### UpdateMethodRuleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rule | [MethodRule](#perm-v1-MethodRule) |  |  |






<a name="perm-v1-UpdateMethodRuleResp"></a>

### UpdateMethodRuleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rule | [MethodRule](#perm-v1-MethodRule) |  |  |





 

 

 


<a name="perm-v1-MethodRuleSrv"></a>

### MethodRuleSrv
menu and function grpc service, MethodRuleSrv provides menu management functions

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Load | [Empty](#perm-v1-Empty) | [BoolResp](#perm-v1-BoolResp) | load method from db, when the menu changes, it needs to be loaded again |
| ListMethodRules | [ListMethodRulesReq](#perm-v1-ListMethodRulesReq) | [ListMethodRulesResp](#perm-v1-ListMethodRulesResp) | list method {platform} is required The returned result is a tree structure |
| CreateMethodRule | [CreateMethodRuleReq](#perm-v1-CreateMethodRuleReq) | [CreateMethodRuleResp](#perm-v1-CreateMethodRuleResp) |  |
| DelMethodRule | [DelMethodRuleReq](#perm-v1-DelMethodRuleReq) | [BoolResp](#perm-v1-BoolResp) |  |
| UpdateMethodRule | [UpdateMethodRuleReq](#perm-v1-UpdateMethodRuleReq) | [BoolResp](#perm-v1-BoolResp) |  |
| GetMethodRule | [GetMethodRuleReq](#perm-v1-GetMethodRuleReq) | [GetMethodRuleResp](#perm-v1-GetMethodRuleResp) |  |

 



<a name="perm_org-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/org.proto



<a name="perm-v1-GetOrgResp"></a>

### GetOrgResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| roles | [GetOrgResp.RolesEntry](#perm-v1-GetOrgResp-RolesEntry) | repeated |  |
| perms | [GetOrgResp.PermsEntry](#perm-v1-GetOrgResp-PermsEntry) | repeated |  |






<a name="perm-v1-GetOrgResp-PermsEntry"></a>

### GetOrgResp.PermsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [List](#perm-v1-List) |  |  |






<a name="perm-v1-GetOrgResp-RolesEntry"></a>

### GetOrgResp.RolesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [List](#perm-v1-List) |  |  |






<a name="perm-v1-ListOrgResp"></a>

### ListOrgResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| orgs | [string](#string) | repeated |  |






<a name="perm-v1-OrgReq"></a>

### OrgReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |






<a name="perm-v1-TransferOrgReq"></a>

### TransferOrgReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| new_user_id | [string](#string) |  |  |





 

 

 


<a name="perm-v1-OrgSrv"></a>

### OrgSrv
org grpc service, OrgSrv provides org info inside RBAC instead of global

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateOrg | [OrgReq](#perm-v1-OrgReq) | [BoolResp](#perm-v1-BoolResp) | CreateOrg init org, create org role and bind all function permissions, the method is idempotent {org_id} is required when {user_id} is set, {user_id} will be admin |
| DelOrg | [OrgReq](#perm-v1-OrgReq) | [BoolResp](#perm-v1-BoolResp) | delete org all perms and data {org_id} is required |
| TransferOrg | [TransferOrgReq](#perm-v1-TransferOrgReq) | [BoolResp](#perm-v1-BoolResp) | transfer org admin to {new_user_id}, then {user_id} will only lose the admin role, and other roles will be retained all parameters are required |
| GetOrg | [OrgReq](#perm-v1-OrgReq) | [GetOrgResp](#perm-v1-GetOrgResp) | get org all perms, for debug |
| ListOrg | [OrgReq](#perm-v1-OrgReq) | [ListOrgResp](#perm-v1-ListOrgResp) | list all org |

 



<a name="perm_perm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/perm.proto



<a name="perm-v1-EnforceReq"></a>

### EnforceReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  | organization |
| user_id | [string](#string) |  | user_id or role_id must be selected |
| role_id | [string](#string) |  |  |
| res_type | [string](#string) |  | resource type, e.g. &#34;box,camera,api etc&#34; |
| node_type | [string](#string) |  |  |
| res_id | [string](#string) |  | resource id, e.g. &#34;box_123,camera_123,/api/cameras,etc&#34; |
| act | [string](#string) |  | resource act, e.g. &#34;post,export,play,download,etc&#34; |






<a name="perm-v1-EnforceResp"></a>

### EnforceResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ok | [bool](#bool) |  |  |
| name | [string](#string) |  |  |






<a name="perm-v1-ListResReq"></a>

### ListResReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| act | [string](#string) |  |  |






<a name="perm-v1-ListResResp"></a>

### ListResResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resources | [ListResResp.ResourcesEntry](#perm-v1-ListResResp-ResourcesEntry) | repeated |  |






<a name="perm-v1-ListResResp-ResourcesEntry"></a>

### ListResResp.ResourcesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [List](#perm-v1-List) |  |  |






<a name="perm-v1-SaveRolePermReq"></a>

### SaveRolePermReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| role_id | [string](#string) |  |  |
| org_id | [string](#string) |  |  |
| method_rules | [string](#string) | repeated |  |
| catalogs | [SaveRolePermReq.Catalog](#perm-v1-SaveRolePermReq-Catalog) | repeated |  |






<a name="perm-v1-SaveRolePermReq-Catalog"></a>

### SaveRolePermReq.Catalog



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is res_id |
| res_type | [string](#string) |  |  |
| node_type | [string](#string) |  |  |





 

 

 


<a name="perm-v1-Perm"></a>

### Perm
Perm grpc service, Perm for authentication and resource list

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Enforce | [EnforceReq](#perm-v1-EnforceReq) | [EnforceResp](#perm-v1-EnforceResp) |  |
| ListResources | [ListResReq](#perm-v1-ListResReq) | [ListResResp](#perm-v1-ListResResp) |  |
| ListMethodRules | [ListMethodRulesReq](#perm-v1-ListMethodRulesReq) | [ListMethodRulesResp](#perm-v1-ListMethodRulesResp) |  |
| ListCatalogs | [ListCatalogsReq](#perm-v1-ListCatalogsReq) | [ListCatalogsResp](#perm-v1-ListCatalogsResp) |  |
| SaveRolePerm | [SaveRolePermReq](#perm-v1-SaveRolePermReq) | [BoolResp](#perm-v1-BoolResp) |  |

 



<a name="perm_role-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## perm/role.proto



<a name="perm-v1-CreateRoleReq"></a>

### CreateRoleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-CreateRoleResp"></a>

### CreateRoleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-DelRoleReq"></a>

### DelRoleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |
| name | [string](#string) |  |  |
| org_id | [string](#string) |  |  |






<a name="perm-v1-DelRoleResp"></a>

### DelRoleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-GetRolePermReq"></a>

### GetRolePermReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  |  |
| org_id | [string](#string) |  |  |






<a name="perm-v1-GetRolePermResp"></a>

### GetRolePermResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| method_rules | [string](#string) | repeated |  |
| catalogs | [GetRolePermResp.CatalogsEntry](#perm-v1-GetRolePermResp-CatalogsEntry) | repeated |  |






<a name="perm-v1-GetRolePermResp-CatalogsEntry"></a>

### GetRolePermResp.CatalogsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [List](#perm-v1-List) |  |  |






<a name="perm-v1-GetRoleReq"></a>

### GetRoleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |






<a name="perm-v1-GetRoleResp"></a>

### GetRoleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-GetRolesForUserResp"></a>

### GetRolesForUserResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| roles | [string](#string) | repeated |  |






<a name="perm-v1-GetUsersForRoleResp"></a>

### GetUsersForRoleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [string](#string) | repeated |  |






<a name="perm-v1-ListRolesReq"></a>

### ListRolesReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |






<a name="perm-v1-ListRolesResp"></a>

### ListRolesResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| roles | [Role](#perm-v1-Role) | repeated |  |






<a name="perm-v1-PermReq"></a>

### PermReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |
| res_type | [string](#string) |  |  |
| node_type | [string](#string) |  |  |
| res_id | [string](#string) |  |  |
| act | [string](#string) |  |  |






<a name="perm-v1-Role"></a>

### Role



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  | role id |
| name | [string](#string) |  | role name, e.g. &#34;admin or 123456&#34; |
| status | [string](#string) |  | role status |
| org_id | [string](#string) |  | org id, |
| display_name | [string](#string) |  | role display name, e.g. &#34;administrators&#34; |
| remark | [string](#string) |  |  |
| created_at | [int64](#int64) |  |  |
| updated_at | [int64](#int64) |  |  |
| deleted_at | [int64](#int64) |  |  |






<a name="perm-v1-UpdateRoleReq"></a>

### UpdateRoleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-UpdateRoleResp"></a>

### UpdateRoleResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#perm-v1-Role) |  |  |






<a name="perm-v1-UserRoleReq"></a>

### UserRoleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| org_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| role_id | [string](#string) |  |  |





 

 

 


<a name="perm-v1-RoleSrv"></a>

### RoleSrv
role grpc service, RoleSrv provides role management and user role management

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateRole | [CreateRoleReq](#perm-v1-CreateRoleReq) | [CreateRoleResp](#perm-v1-CreateRoleResp) | create role |
| DelRole | [DelRoleReq](#perm-v1-DelRoleReq) | [BoolResp](#perm-v1-BoolResp) | - delete role by id or name - req: id=12 - req: name=admin,org_id=ka |
| UpdateRole | [UpdateRoleReq](#perm-v1-UpdateRoleReq) | [BoolResp](#perm-v1-BoolResp) | - update role by id or name - req: id=12 - req: name=admin,org_id=ka |
| GetRole | [GetRoleReq](#perm-v1-GetRoleReq) | [GetRoleResp](#perm-v1-GetRoleResp) | get role by id |
| ListRoles | [ListRolesReq](#perm-v1-ListRolesReq) | [ListRolesResp](#perm-v1-ListRolesResp) | list role by org_id |
| AddRoleForUser | [UserRoleReq](#perm-v1-UserRoleReq) | [BoolResp](#perm-v1-BoolResp) | add role to user all parameters are required |
| DelRoleForUser | [UserRoleReq](#perm-v1-UserRoleReq) | [BoolResp](#perm-v1-BoolResp) | delete user org all parameters are required if {role_id} is *, it will delete all role about the user |
| GetRolesForUser | [UserRoleReq](#perm-v1-UserRoleReq) | [GetRolesForUserResp](#perm-v1-GetRolesForUserResp) | get user all roles {org_id} and {user_id} are required |
| GetUsersForRole | [UserRoleReq](#perm-v1-UserRoleReq) | [GetUsersForRoleResp](#perm-v1-GetUsersForRoleResp) | get users from a {role_id} if {role_id} is null, you will get all users of the {org_id} |

 



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

