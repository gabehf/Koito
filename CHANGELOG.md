# v0.0.10

## Features
- Support for custom themes added! You can find the custom theme input in the Appearance menu.
- Allow loading environment variables from files using the _FILE suffix (#20)
- All activity grids (calendar heatmaps) are now configurable

## Enhancements
- The activity grid on the home page is now configurable

## Fixes
- Sub-second precision is stripped from incoming listens to ensure they can be deleted reliably
- Top items are now sorted by id for stability

## Updates
- Adjusted colors for the "Yuu" theme
- Themes now have a single source of truth in themes.css.ts
- Configurable activity grids now have a re-styled, collapsible menu
- The year option for activity grids has been removed