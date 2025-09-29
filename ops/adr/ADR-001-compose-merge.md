# ADR-001 Compose Merge
- Prefix names: {namespace}__{name}, map overrides allowed.
- Rewrite depends_on.
- Var substitution: project > module default.
- Port collision policy: error (default) or auto-rewrite(disabled for now).
- Output: merged yaml + mapping table.
