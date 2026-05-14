{{ config(
    materialized='view'
) }}

WITH base AS (
    SELECT
    *
    FROM {{ source('raw_sales', 'customers') }}
)

SELECT
    base.customer_id AS customer_id,
    base.full_name AS full_name,
    base.email AS email
FROM base
