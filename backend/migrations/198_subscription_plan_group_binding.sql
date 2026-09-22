-- Subscription plans can be explicitly detached from their routing group.
-- group_id is retained as the last/selected target for compatibility, while
-- group_bound controls whether the relationship is active.
ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS group_bound BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_subscription_plans_group_bound
    ON subscription_plans(group_bound);

-- A detached plan must never remain purchasable.
UPDATE subscription_plans
SET for_sale = FALSE
WHERE group_bound = FALSE;
