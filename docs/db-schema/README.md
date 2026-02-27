# billing

## Tables

| Name | Columns | Comment | Type |
| ---- | ------- | ------- | ---- |
| [public.tier](public.tier.md) | 3 |  | BASE TABLE |
| [public.customer](public.customer.md) | 6 |  | BASE TABLE |
| [public.cf_org](public.cf_org.md) | 3 |  | BASE TABLE |
| [public.meter](public.meter.md) | 1 | A Meter reads usage information from a system in Cloud.gov. It also namespaces natural IDs for resources and resource_kinds; meter + natural_id is a primary key. | BASE TABLE |
| [public.resource_kind](public.resource_kind.md) | 3 | ResourceKind represents a particular kind of billable resource. Note that natural_id can be empty because some meters may only read one kind of resource, and that resource kind may not have a unique identifier in the target system; it is uniquely identified by the meter name only. | BASE TABLE |
| [public.resource](public.resource.md) | 4 |  | BASE TABLE |
| [public.reading](public.reading.md) | 4 |  | BASE TABLE |
| [public.measurement](public.measurement.md) | 7 |  | BASE TABLE |
| [public.account_type](public.account_type.md) | 3 |  | BASE TABLE |
| [public.account](public.account.md) | 3 |  | BASE TABLE |
| [public.transaction](public.transaction.md) | 5 |  | BASE TABLE |
| [public.entry](public.entry.md) | 4 |  | BASE TABLE |
| [public.price](public.price.md) | 7 |  | BASE TABLE |
| [public.resource_node](public.resource_node.md) | 4 |  | BASE TABLE |

## Stored procedures and functions

| Name | ReturnType | Arguments | Type |
| ---- | ------- | ------- | ---- |
| public.assert_transaction_balances | trigger |  | FUNCTION |
| public.bounds_month_prev | record | as_of timestamp with time zone DEFAULT now(), tz text DEFAULT 'America/New_York'::text | FUNCTION |
| public.update_measurement_microcredits | int8 | as_of timestamp with time zone DEFAULT now() | FUNCTION |
| public.post_usage | int4 | as_of timestamp with time zone DEFAULT now() | FUNCTION |
| public.uuid_generate_v7 | uuid |  | FUNCTION |
| public.ltree_in | ltree | cstring | FUNCTION |
| public.ltree_out | cstring | ltree | FUNCTION |
| public.ltree_cmp | int4 | ltree, ltree | FUNCTION |
| public.ltree_lt | bool | ltree, ltree | FUNCTION |
| public.ltree_le | bool | ltree, ltree | FUNCTION |
| public.ltree_eq | bool | ltree, ltree | FUNCTION |
| public.ltree_ge | bool | ltree, ltree | FUNCTION |
| public.ltree_gt | bool | ltree, ltree | FUNCTION |
| public.ltree_ne | bool | ltree, ltree | FUNCTION |
| public.subltree | ltree | ltree, integer, integer | FUNCTION |
| public.subpath | ltree | ltree, integer, integer | FUNCTION |
| public.subpath | ltree | ltree, integer | FUNCTION |
| public.index | int4 | ltree, ltree | FUNCTION |
| public.index | int4 | ltree, ltree, integer | FUNCTION |
| public.nlevel | int4 | ltree | FUNCTION |
| public.ltree2text | text | ltree | FUNCTION |
| public.text2ltree | ltree | text | FUNCTION |
| public.lca | ltree | ltree[] | FUNCTION |
| public.lca | ltree | ltree, ltree | FUNCTION |
| public.lca | ltree | ltree, ltree, ltree | FUNCTION |
| public.lca | ltree | ltree, ltree, ltree, ltree | FUNCTION |
| public.lca | ltree | ltree, ltree, ltree, ltree, ltree | FUNCTION |
| public.lca | ltree | ltree, ltree, ltree, ltree, ltree, ltree | FUNCTION |
| public.lca | ltree | ltree, ltree, ltree, ltree, ltree, ltree, ltree | FUNCTION |
| public.lca | ltree | ltree, ltree, ltree, ltree, ltree, ltree, ltree, ltree | FUNCTION |
| public.ltree_isparent | bool | ltree, ltree | FUNCTION |
| public.ltree_risparent | bool | ltree, ltree | FUNCTION |
| public.ltree_addltree | ltree | ltree, ltree | FUNCTION |
| public.ltree_addtext | ltree | ltree, text | FUNCTION |
| public.ltree_textadd | ltree | text, ltree | FUNCTION |
| public.ltreeparentsel | float8 | internal, oid, internal, integer | FUNCTION |
| public.lquery_in | lquery | cstring | FUNCTION |
| public.lquery_out | cstring | lquery | FUNCTION |
| public.ltq_regex | bool | ltree, lquery | FUNCTION |
| public.ltq_rregex | bool | lquery, ltree | FUNCTION |
| public.lt_q_regex | bool | ltree, lquery[] | FUNCTION |
| public.lt_q_rregex | bool | lquery[], ltree | FUNCTION |
| public.ltxtq_in | ltxtquery | cstring | FUNCTION |
| public.ltxtq_out | cstring | ltxtquery | FUNCTION |
| public.ltxtq_exec | bool | ltree, ltxtquery | FUNCTION |
| public.ltxtq_rexec | bool | ltxtquery, ltree | FUNCTION |
| public.ltree_gist_in | ltree_gist | cstring | FUNCTION |
| public.ltree_gist_out | cstring | ltree_gist | FUNCTION |
| public.ltree_consistent | bool | internal, ltree, smallint, oid, internal | FUNCTION |
| public.ltree_compress | internal | internal | FUNCTION |
| public.ltree_decompress | internal | internal | FUNCTION |
| public.ltree_penalty | internal | internal, internal, internal | FUNCTION |
| public.ltree_picksplit | internal | internal, internal | FUNCTION |
| public.ltree_union | ltree_gist | internal, internal | FUNCTION |
| public.ltree_same | internal | ltree_gist, ltree_gist, internal | FUNCTION |
| public._ltree_isparent | bool | ltree[], ltree | FUNCTION |
| public._ltree_r_isparent | bool | ltree, ltree[] | FUNCTION |
| public._ltree_risparent | bool | ltree[], ltree | FUNCTION |
| public._ltree_r_risparent | bool | ltree, ltree[] | FUNCTION |
| public._ltq_regex | bool | ltree[], lquery | FUNCTION |
| public._ltq_rregex | bool | lquery, ltree[] | FUNCTION |
| public._lt_q_regex | bool | ltree[], lquery[] | FUNCTION |
| public._lt_q_rregex | bool | lquery[], ltree[] | FUNCTION |
| public._ltxtq_exec | bool | ltree[], ltxtquery | FUNCTION |
| public._ltxtq_rexec | bool | ltxtquery, ltree[] | FUNCTION |
| public._ltree_extract_isparent | ltree | ltree[], ltree | FUNCTION |
| public._ltree_extract_risparent | ltree | ltree[], ltree | FUNCTION |
| public._ltq_extract_regex | ltree | ltree[], lquery | FUNCTION |
| public._ltxtq_extract_exec | ltree | ltree[], ltxtquery | FUNCTION |
| public._ltree_consistent | bool | internal, ltree[], smallint, oid, internal | FUNCTION |
| public._ltree_compress | internal | internal | FUNCTION |
| public._ltree_penalty | internal | internal, internal, internal | FUNCTION |
| public._ltree_picksplit | internal | internal, internal | FUNCTION |
| public._ltree_union | ltree_gist | internal, internal | FUNCTION |
| public._ltree_same | internal | ltree_gist, ltree_gist, internal | FUNCTION |
| public.ltree_recv | ltree | internal | FUNCTION |
| public.ltree_send | bytea | ltree | FUNCTION |
| public.lquery_recv | lquery | internal | FUNCTION |
| public.lquery_send | bytea | lquery | FUNCTION |
| public.ltxtq_recv | ltxtquery | internal | FUNCTION |
| public.ltxtq_send | bytea | ltxtquery | FUNCTION |
| public.ltree_gist_options | void | internal | FUNCTION |
| public._ltree_gist_options | void | internal | FUNCTION |
| public.move_customer | bool | p_customer_id uuid, p_new_parent_id uuid | FUNCTION |
| public.move_resource_node | bool | p_resource_node_id uuid, p_new_parent_id uuid | FUNCTION |

## Enums

| Name | Values |
| ---- | ------- |
| public.river_job_state | available, cancelled, completed, discarded, pending, retryable, running, scheduled |
| public.transaction_type | iaa_pop_end, iaa_pop_start, usage_post |

## Relations

```mermaid
erDiagram

"public.customer" }o--o| "public.tier" : "FOREIGN KEY (tier_id) REFERENCES tier(id)"
"public.cf_org" }o--o| "public.customer" : "FOREIGN KEY (customer_id) REFERENCES customer(id)"
"public.resource_kind" }o--|| "public.meter" : "FOREIGN KEY (meter) REFERENCES meter(name)"
"public.resource" }o--|| "public.cf_org" : "FOREIGN KEY (cf_org_id) REFERENCES cf_org(id)"
"public.resource" }o--|| "public.resource_kind" : "FOREIGN KEY (meter, kind_natural_id) REFERENCES resource_kind(meter, natural_id)"
"public.measurement" }o--|| "public.resource" : "FOREIGN KEY (meter, resource_natural_id) REFERENCES resource(meter, natural_id)"
"public.measurement" }o--|| "public.reading" : "FOREIGN KEY (reading_id) REFERENCES reading(id)"
"public.measurement" }o--o| "public.transaction" : "FOREIGN KEY (transaction_id) REFERENCES transaction(id)"
"public.measurement" }o--o| "public.price" : "FOREIGN KEY (price_id) REFERENCES price(id)"
"public.account" }o--|| "public.account_type" : "FOREIGN KEY (type) REFERENCES account_type(id)"
"public.account" }o--o| "public.customer" : "FOREIGN KEY (customer_id) REFERENCES customer(id)"
"public.entry" }o--|| "public.account" : "FOREIGN KEY (account_id) REFERENCES account(id)"
"public.entry" }o--|| "public.transaction" : "FOREIGN KEY (transaction_id) REFERENCES transaction(id)"
"public.price" }o--|| "public.resource_kind" : "FOREIGN KEY (meter, kind_natural_id) REFERENCES resource_kind(meter, natural_id)"
"public.resource_node" }o--|| "public.customer" : "FOREIGN KEY (customer_id) REFERENCES customer(id)"

"public.tier" {
  integer id
  text name
  bigint tier_credits
}
"public.customer" {
  bigint old_id
  text name
  integer tier_id FK
  uuid id
  ltree path
  varchar_256_ slug
}
"public.cf_org" {
  uuid id
  text name
  uuid customer_id FK
}
"public.meter" {
  text name
}
"public.resource_kind" {
  text meter FK
  text natural_id
  text name
}
"public.resource" {
  text meter FK
  text natural_id
  text kind_natural_id FK
  uuid cf_org_id FK
}
"public.reading" {
  integer id
  timestamp_without_time_zone created_at
  boolean periodic
  timestamp_with_time_zone created_at_utc
}
"public.measurement" {
  integer reading_id FK
  text meter FK
  text resource_natural_id FK
  integer value
  bigint amount_microcredits
  bigint transaction_id FK
  bigint price_id FK
}
"public.account_type" {
  integer id
  text name
  integer normal
}
"public.account" {
  integer id
  integer type FK
  uuid customer_id FK
}
"public.transaction" {
  integer id
  timestamp_with_time_zone occurred_at
  text description
  transaction_type type
  uuid customer_id
}
"public.entry" {
  integer transaction_id FK
  integer account_id FK
  integer direction
  bigint amount_microcredits
}
"public.price" {
  integer id
  text meter FK
  text kind_natural_id FK
  text unit_of_measure
  bigint microcredits_per_unit
  bigint unit
  tstzrange valid_during
}
"public.resource_node" {
  ltree path
  varchar_256_ slug
  uuid customer_id FK
  text resource_natural_id
}
```

---

> Generated by [tbls](https://github.com/k1LoW/tbls)
