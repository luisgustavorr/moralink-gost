# MoraLink Field Mapper DSL

Reference documentation for the JSON-based field mapping language used to configure API integrations in MoraLink.

---

## Overview

Each API integration is configured via a JSON document that tells the mapper how to fetch data, how to traverse the response, and how to translate fields into MoraLink's internal format.

The document has the following top-level keys:

```json
{
  "url": "...",
  "id_1": { ... },
  "id_2": { ... },
  "id_3": { ... },
  "individual_detail": { ... },
  "fields": [ ... ]
}
```

| Key | Description |
|-----|-------------|
| `url` | Base URL for the API request |
| `id_N` | URL parameter injectors (dynamic query string values); `id_1`, `id_2`, and `id_3` are supported |
| `individual_detail` | Optional config for a secondary per-record fetch (see [Individual Detail](#individual-detail-individual_detail)) |
| `fields` | Ordered list of field mapping rules |

---

## URL Parameter Injectors (`id_N`)

Inject dynamic values into the API request URL at call time. `id_1`, `id_2`, and `id_3` are all available; each integation function decides which ones it reads and how it concatenates them into the final URL.

```json
"id_1": {
  "key": "?dataInicio=",
  "value": "<value_expression>"
}
```

| Property | Type | Description |
|----------|------|-------------|
| `key` | string | Query string key including separator (`?` or `&`) or a path segment |
| `value` | string | Value expression (see below) |

### Value Expressions

| Pattern | Description | Example |
|---------|-------------|---------|
| `days_ago!N!format` | Date N days in the past, formatted using Go time layout | `days_ago!10!02/01/2006` |
| `token` | Replaced at runtime with the current bearer token | `"token"` |
| Any other literal string | Written as-is into the URL | `"ativo"` |

**Go date format reference:** `02/01/2006` = DD/MM/YYYY, `2006-01-02` = YYYY-MM-DD.

---

## Individual Detail (`individual_detail`)

When present, the mapper makes a secondary HTTP GET request for every element before applying the `fields` rules. The response from that sub-request is stored internally and can be activated for specific field rules via `switch_to_details: true`.

```json
"individual_detail": {
  "url": "https://api.example.com/clientes/{id}",
  "key_getter": {
    "src": "id",
    "dst": "{id}"
  },
  "id_1": {
    "key": "?token=",
    "value": "token"
  }
}
```

| Property | Type | Description |
|----------|------|-------------|
| `url` | string | URL template; `{dst}` placeholders are replaced with the value resolved from `key_getter.src` in the current element |
| `key_getter.src` | string | Dot-path into the current element used to build the URL |
| `key_getter.dst` | string | Placeholder string in the URL template to replace |
| `id_1` | Id object | Optional extra query-string parameter appended after the resolved URL |

To use the individual detail response instead of the main element for a given field, set `switch_to_details: true` on that rule. Once switched, all subsequent rules in the list also operate on the detail response.

```json
{ "switch_to_details": true, "src": "fullName", "dst": "nome" }
```

---

## Field Rules (`fields[]`)

Each object in `fields` maps one value from the raw API response to one destination key in MoraLink's data model. Rules are applied in order.

### Common Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `dst` | string | ✅ | Destination key — must match the `db` struct tag on the target row type |
| `src` | string | — | Source key or dot-path in the current element |
| `op` | string | — | Operation to apply (see [Operations](#operations)) |
| `src_raw_value` | string | — | Write a literal value; no source lookup performed |
| `src_list` | string[] | — | Coalesce — additional fallback keys tried after `src` |
| `src_object_builder` | object | — | Build a nested JSON object (triggers `build_object` behaviour) |
| `src_payment_status` | object | — | Resolve a payment status string (triggers payment status logic) |
| `nullif` | string | — | Treat the resolved value as empty if it equals this string |
| `format_date` | object | — | Date format config when `op` is `format_date` |
| `duration_rules` | object | — | Category-to-duration map when `op` is `calc_duration` |
| `case` | object | — | Conditional value mapping when `op` is `case` |
| `switch_to_details` | bool | — | Switch the data source to the individual detail response for this and all subsequent rules |

---

## Source Types

### 1. Simple key

Reads a top-level key from the current element as a string.

```json
{ "src": "valorAtualizado", "dst": "valor_total" }
```

---

### 2. Dot-path traversal — `op: "extract"`

Traverses nested objects and arrays using `.` as separator.

```json
{ "src": "cliente.id", "dst": "cliente", "op": "extract" }
```

Array index access uses the numeric index as a path segment:

```json
{ "src": "itens.0.valor", "dst": "primeiro_valor", "op": "extract" }
```

---

### 3. Raw literal — `src_raw_value`

Writes a constant string directly to the destination. No source lookup is performed. Takes precedence over all other source types.

```json
{ "src_raw_value": "1", "dst": "ativo" }
{ "src_raw_value": "0", "dst": "data_personalizadas" }
```

---

### 4. Coalesce — `src_list`

Tries `src` first, then each key in `src_list` in order, and uses the first non-empty value. Combined with `nullif`, values equal to the nullif string are skipped and the next candidate is tried.

```json
{
  "src": "preco05",
  "src_list": ["preco00", "preco01"],
  "dst": "valor",
  "nullif": "0"
}
```

> Note: `src_list` is checked against the raw element using the same dot-path resolution as `op: "extract"`, so nested paths work in the fallback list too.

---

### 5. Payment status — `src_payment_status`

Resolves a payment status string by inspecting two date fields: one for the expiry/due date and one for the paid date. The resolved value is one of three strings: `"paga"`, `"criada"`, or `"vencida"`.

Resolution logic:

1. If the expire source field is `nil` → `"criada"`
2. If the paid source field is non-nil → `"paga"`
3. If today is before the expire date → `"criada"`
4. Otherwise → `"vencida"`

```json
{
  "src_payment_status": {
    "expire": {
      "src": "vencimento",
      "format_date": { "raw": "02/01/2006", "dst": "2006-01-02" }
    },
    "paid": {
      "src": "recebimento",
      "format_date": { "raw": "02/01/2006", "dst": "2006-01-02" }
    }
  },
  "dst": "status"
}
```

| Sub-key | Description |
|---------|-------------|
| `expire` | Which source field holds the due date, and how to parse it |
| `paid` | Which source field holds the payment date; presence (non-nil) means paid |
| `format_date` | Date parsing config applied to each source field before comparison |

> The `format_date` inside `src_payment_status` is used only for parsing the date to compare against today; the output written to `dst` is always one of the three status strings above, never a formatted date.

---

### 6. Object builder — `src_object_builder`

Builds a nested object (or array of objects) by running a recursive sub-transcriptor on the current element or on a sub-array within it. The result is JSON-serialised and stored in the destination field (which should map to a `*[]byte` column).

```json
{
  "src_object_builder": {
    "get_from": "",
    "object_builder": {
      "fields": [
        { "src": "vencimento",      "dst": "data_vencimento" },
        { "src": "valorAtualizado", "dst": "valor_parcela" },
        { "src": "linkBoleto",      "dst": "id_boleto" },
        { "src_raw_value": "ativo", "dst": "status" }
      ]
    }
  },
  "dst": "infos_cobranca",
  "op": "build_object"
}
```

| Property | Description |
|----------|-------------|
| `get_from` | Dot-path to the source array inside the current element. Empty string `""` wraps the current element itself in a single-element slice |
| `object_builder.fields` | A full nested `fields` array — supports all the same rule types recursively |

> `op: "build_object"` must be set alongside `src_object_builder` for the rule to be recognised.

---

## Operations (`op`)

The `op` field controls how the resolved `src` value is transformed before being written to `dst`.

| Value | Description |
|-------|-------------|
| *(omitted)* | Plain map key lookup — `dst = element[src]` |
| `extract` | Dot-path traversal into nested objects or arrays |
| `format_date` | Parse date from `format_date.raw` layout and output in `format_date.dst` layout |
| `build_object` | Build a nested JSON object using `src_object_builder` |
| `calc_duration` | Map the resolved value to a duration string using `duration_rules` |
| `case` | Conditional value mapping using `case.conditions` and `case.default` |

> Operations `disc_percent`, `disc_value`, `coalesce`, and `fetch` appear in internal constants but are not implemented in the current transcription switch. Do not use them in configs.

---

## `format_date` Config

Parses a date string from one format and outputs it in another. Used as a standalone `op` or embedded inside `src_payment_status`.

```json
{
  "src": "vencimento",
  "dst": "data_vencimento",
  "op": "format_date",
  "format_date": {
    "raw": "02/01/2006",
    "dst": "2006-01-02"
  }
}
```

| Property | Description |
|----------|-------------|
| `raw` | Input date format using Go time layout |
| `dst` | Output date format using Go time layout |

If the source value is an empty string, an empty string is written to the destination without error.

**Common Go layout tokens:**

| Token | Meaning |
|-------|---------|
| `2006` | Year (4-digit) |
| `01` | Month (zero-padded) |
| `02` | Day (zero-padded) |
| `15` | Hour 24h |
| `04` | Minute |
| `05` | Second |

---

## `calc_duration` Config

Maps a source value (typically a category ID) to a duration string by looking it up in a rules map. The keys are the output duration values; each key maps to a list of source values that produce that duration. If no match is found, the result is `"0"`.

```json
{
  "src": "categoriaProduto.id",
  "dst": "duracao",
  "op": "calc_duration",
  "duration_rules": {
    "0":   ["1002", "112003", "18"],
    "30":  ["1004", "108003", "5", "6"],
    "60":  ["2"],
    "365": ["1", "3"]
  }
}
```

The `src` field supports dot-path traversal (resolved via `ResolvePath`), so nested category objects work directly.

---

## `case` Config

Evaluates a list of conditions against the resolved `src` value and writes the first matching `then` value to `dst`. If no condition matches, `default` is written.

```json
{
  "src": "situacao",
  "dst": "status",
  "op": "case",
  "case": {
    "conditions": [
      { "when": "A", "then": "ativo" },
      { "when": "I", "then": "inativo" },
      { "when": "P", "then": "pendente" }
    ],
    "default": "desconhecido"
  }
}
```

| Property | Description |
|----------|-------------|
| `conditions` | Ordered list of `{ "when": <string>, "then": <string> }` pairs |
| `default` | Value written when no condition matches |

Matching is done by strict string equality between `when` and the value returned by `ResolvePath(element, src)`. Only the first matching condition is applied.

---

## Full Example

```json
{
  "url": "https://api.example.com/financeiro",
  "id_1": {
    "key": "?dataInicio=",
    "value": "days_ago!10!02/01/2006"
  },
  "fields": [
    { "src": "idParcelaReceita", "dst": "id_externo" },
    { "src": "cliente.id",       "dst": "cliente",   "op": "extract" },
    {
      "src_payment_status": {
        "expire": { "src": "vencimento",  "format_date": { "raw": "02/01/2006", "dst": "2006-01-02" } },
        "paid":   { "src": "recebimento", "format_date": { "raw": "02/01/2006", "dst": "2006-01-02" } }
      },
      "dst": "status"
    },
    { "src": "valorAtualizado",         "dst": "valor_total" },
    { "src": "movimento.quantParcelas", "dst": "parcelas", "op": "extract" },
    { "src": "valorAtualizado",         "dst": "valor_parcela" },
    { "src_raw_value": "1",             "dst": "ativo" },
    {
      "src": "vencimento",
      "dst": "data_vencimento",
      "op": "format_date",
      "format_date": { "raw": "02/01/2006", "dst": "2006-01-02" }
    },
    { "src_raw_value": "0", "dst": "data_personalizadas" },
    {
      "src": "categoriaProduto.id",
      "dst": "duracao",
      "op": "calc_duration",
      "duration_rules": {
        "0":   ["1002", "18"],
        "30":  ["1004", "5", "6"],
        "365": ["1", "3"]
      }
    },
    {
      "src": "situacao",
      "dst": "status",
      "op": "case",
      "case": {
        "conditions": [
          { "when": "A", "then": "ativo" },
          { "when": "I", "then": "inativo" }
        ],
        "default": "desconhecido"
      }
    },
    {
      "src_object_builder": {
        "get_from": "",
        "object_builder": {
          "fields": [
            { "src": "vencimento",      "dst": "data_vencimento" },
            { "src": "valorAtualizado", "dst": "valor_parcela" },
            { "src": "linkBoleto",      "dst": "id_boleto" },
            { "src_raw_value": "ativo", "dst": "status" }
          ]
        }
      },
      "dst": "infos_cobranca",
      "op": "build_object"
    }
  ]
}
```

---

## Rule Evaluation Order

For each element, `Transcribe` processes rules in array order. Within a single rule, the source is resolved in the following priority:

1. `switch_to_details: true` — swap the active element to the individual detail response
2. `src_payment_status` — payment status logic; writes result and moves to next rule
3. `src_object_builder` — object builder; writes result and moves to next rule
4. `src_raw_value` — literal write; moves to next rule
5. `src_list` coalesce — override `src` with the first non-empty fallback
6. `op` switch — apply the chosen operation to the resolved `src`

---

## Where Configs Live

Transcriptor JSONs are stored in the database under each integration's configuration record and loaded at sync time via `JsonToTranscriptor([]byte)`. They are associated with a specific API manager (`frontsys.go`, `tray.go`, etc.) and a specific table (`produtos`, `financeiro`, `vendas`, etc.).