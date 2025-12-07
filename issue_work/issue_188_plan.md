# Issue 188 Implementation Plan

## Summary
Replace the placeholder telemetry page with a full privacy policy page that explains what data tsuku collects, what it does not collect, and how to opt out.

## Approach
Direct modification of the existing `website/telemetry/index.html` placeholder. The page will use the existing dark theme CSS variables and follow the established content section patterns.

### Alternatives Considered
- **Create a new page structure**: Not needed since the placeholder already has the correct header/nav/footer structure
- **Use markdown with a build step**: Would add complexity; static HTML is preferred per project conventions

## Files to Modify
- `website/telemetry/index.html` - Replace placeholder content with full privacy policy
- `website/assets/style.css` - Add styles for tables and structured content sections

## Files to Create
None

## Implementation Steps
- [ ] Add CSS styles for privacy page tables and content layout
- [ ] Replace placeholder HTML with complete privacy policy content
- [ ] Verify page renders correctly and is mobile-responsive

## Testing Strategy
- Manual verification: View page locally with `python3 -m http.server 8000`
- Check mobile responsiveness using browser dev tools
- Verify all links work (back to home, stats page, GitHub links)

## Risks and Mitigations
- **Outdated telemetry info**: Content is based on issue #188 which has the authoritative field list
- **CSS conflicts**: New table styles will be scoped appropriately to avoid affecting other pages

## Success Criteria
- [ ] Page accessible at `/telemetry`
- [ ] Lists all collected fields with explanations (recipe, version, os, arch, tsuku_version, action)
- [ ] Clearly states what is NOT collected (IP, identifiers, hostnames, etc.)
- [ ] Provides opt-out instructions (TSUKU_NO_TELEMETRY=1, --no-telemetry flag)
- [ ] Data retention policy included (90 days raw, indefinite aggregated)
- [ ] Source code links included
- [ ] Mobile-responsive
- [ ] Consistent styling with main site

## Open Questions
None - all requirements are clearly defined in issue #188
