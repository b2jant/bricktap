MODEL (
  name enriched_orders,
  kind FULL,
  description 'Cleaned orders unnested with customer details attached.'
);

WITH base AS (
    SELECT
    *
    FROM raw_stripe.payments
),

customer AS (
    SELECT
    *
    FROM core.customers
)

SELECT
    NULLIF(TRIM(id), '') AS user_id,
    customer.email AS customer_email,
    base.login AS total_logins,
    MIN(created_at) OVER (PARTITION BY user_id) AS first_active_date
FROM base
LEFT JOIN customer
  ON base.user_id = customer.user_id
-- WARNING: Unnest is not supported in the Base ANSI dialect. Please specify a target data warehouse.
WHERE base.status = 'active'
-- WARNING: Pivot is not supported in the Base ANSI dialect. Please specify a target data warehouse.
