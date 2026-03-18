# MoraLink Field Mapper DSL

Reference documentation for the JSON-based field mapping language used to configure API integrations in MoraLink.

---

## Overview

Each API integration is configured via a JSON document that tells the mapper how to fetch data, how to traverse the response, and how to translate fields into MoraLink's internal format.

The document has two top-level sections:

```json
{
  "id_1": { ... },
  "id_2": { ... },
  "fields": [ ... ]
}
```

| Key | Description |
|-----|-------------|
| `id_N` | URL parameter injectors (dynamic query string values) |
| `fields` | Ordered list of field mapping rules |

---

## URL Parameter Injectors (`id_N`)

Inject dynamic values into the API request URL at call time.

```json
"id_1": {
  "key": "?dataInicio=",
  "value": "<value_expression>"
}
```

| Property | Type | Description |
|----------|------|-------------|
| `key` | string | Query string key including separator (`?` or `&`) |
| `value` | string | Value expression (see below) |

### Value Expressions

| Pattern | Description | Example |
|---------|-------------|---------|
| `days_ago!N!format` | Date N days in the past, formatted using Go time layout | `days_ago!10!02/01/2006` |
| Any literal string | Written as-is into the URL | `"ativo"` |

**Go date format reference:** `02/01/2006` = DD/MM/YYYY, `2006-01-02` = YYYY-MM-DD.

---

## Field Rules (`fields[]`)

Each object in `fields` maps one value from the raw API response to one destination key in MoraLink's data model. Rules are applied in order.

### Common Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `dst` | string | ✅ | Destination key — must match the `db` struct tag |
| `src` | string | — | Source key or dot-path in the raw element |
| `op` | string | — | Operation to apply (see [Operations](#operations)) |
| `src_raw_value` | string | — | Write a literal value; no source lookup performed |
| `src_list` | string[] | — | Coalesce — tries each key until one is non-empty |
| `nullif` | string | — | Treat the resolved value as empty if it equals this string |
| `format_date` | object | — | Date format config when `op` is `format_date` |
| `duration_rules` | object | — | Category-to-duration map when `op` is `calc_duration` |

---

## Source Types

### 1. Simple key

Reads a top-level key from the raw element as a string.

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

Writes a constant string directly to the destination. No source lookup is performed.

```json
{ "src_raw_value": "1", "dst": "ativo" }
{ "src_raw_value": "0", "dst": "data_personalizadas" }
```

---

### 4. Coalesce — `src_list`

Tries each key in `src_list` in order and uses the first non-empty value. `src` is the primary candidate; `src_list` are fallbacks.

```json
{
  "src": "preco05",
  "src_list": ["preco00", "preco01"],
  "dst": "valor",
  "nullif": "0"
}
```

Combined with `nullif`, values equal to the nullif string are skipped and the next candidate is tried.

---

### 5. Payment status — `src_payment_status`

Resolves a payment status field based on which date field is populated in the source. If `paid.src` has a value, the record is considered paid; otherwise it is considered pending/expired.

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
| `expire` | Configuration for the pending/expired path — which source field to read and how to format its date |
| `paid` | Configuration for the paid path — which source field to read and how to format its date |
| `format_date` | Optional date reformatting applied to the resolved value (see [format_date](#format_date-config)) |

---

### 6. Object builder — `op: "build_object"`

Builds a nested object (or array of objects) by running a recursive sub-transcriptor on the current element or a sub-array within it.

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
| `get_from` | Dot-path to the source array inside the current element. Empty string `""` means use the current element itself |
| `object_builder.fields` | A full nested `fields` array — supports all the same rule types recursively |

The result is serialized to JSON and stored in the destination field (which maps to a `*[]byte` / `db:"..."` raw column).

---

## Operations (`op`)

| Value | Description |
|-------|-------------|
| *(omitted)* | Plain string copy from `src` |
| `extract` | Dot-path traversal into nested object or array |
| `format_date` | Parse date from `raw` format and output in `dst` format |
| `build_object` | Build a nested object using a sub-transcriptor (requires `src_object_builder`) |
| `calc_duration` | Map a category ID to a duration string using `duration_rules` |
| `disc_percent` | Apply a percentage discount to the source value |
| `disc_value` | Apply a fixed value discount to the source value |
| `coalesce` | Use `src_list` to find the first non-empty value |
| `fetch` | Make a sub-request to an external endpoint and extract a value from the response |

---

## `format_date` Config

Used as a standalone `op` or inline inside `src_payment_status`.

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

Maps a source value (typically a category ID) to a duration string by looking it up in a rules map.

```json
{
  "src": "categoriaProduto.id",
  "dst": "duracao",
  "op": "calc_duration",
  "duration_rules": {
    "0":  ["1002", "112003", "18"],
    "30": ["1004", "108003", "5", "6"],
    "60": ["2"],
    "365": ["1", "3"]
  }
}
```

The keys are the output duration values. The arrays are the source IDs that map to each duration. If no match is found, the result is `"0"`.

---

## `fetch` Config

Makes a secondary HTTP request during transcription and extracts a value from its response.

```json
{
  "src": "/clientes/{id}",
  "dst": "cliente_info",
  "op": "fetch",
  "method": "GET",
  "extract": "data.items",
  "alias": "id->id_externo,nome->name"
}
```

| Property | Description |
|----------|-------------|
| `src` | URL template — `{key}` placeholders are replaced with values from the current element |
| `method` | HTTP method (`GET`, `POST`, etc.) |
| `extract` | Dot-path into the fetch response to extract the desired value |
| `alias` | Comma-separated `src_key->dst_key` pairs for remapping keys in the extracted result |

---

## Full Example

```json
{
  "id_1": {
    "key": "?dataInicio=",
    "value": "days_ago!10!02/01/2006"
  },
  "fields": [
    { "src": "idParcelaReceita", "dst": "id_externo" },
    { "src": "cliente.id",       "dst": "cliente", "op": "extract" },
    {
      "src_payment_status": {
        "expire": { "src": "vencimento",   "format_date": { "raw": "02/01/2006", "dst": "2006-01-02" } },
        "paid":   { "src": "recebimento",  "format_date": { "raw": "02/01/2006", "dst": "2006-01-02" } }
      },
      "dst": "status"
    },
    { "src": "valorAtualizado",          "dst": "valor_total" },
    { "src": "movimento.quantParcelas",  "dst": "parcelas", "op": "extract" },
    { "src": "valorAtualizado",          "dst": "valor_parcela" },
    { "src_raw_value": "1",              "dst": "ativo" },
    {
      "src": "vencimento",
      "dst": "data_vencimento",
      "op": "format_date",
      "format_date": { "raw": "02/01/2006", "dst": "2006-01-02" }
    },
    { "src_raw_value": "0", "dst": "data_personalizadas" },
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

## Where Configs Live

Transcriptor JSONs are stored in the database under each integration's configuration record and loaded at sync time via `JsonToTranscriptor([]byte)`. They are associated with a specific API manager (`frontsys.go`, `shark.go`, etc.) and a specific table (`produtos`, `financeiro`, `vendas`, etc.).