# Submission (Meta Global AI Developer Hackathon, due Nov 4, 2026, 9:00 PM ET)

Track: Indie developer. Theme: Build for Your World.

## Project name
Rehearsal

## One line
Stress-test a small brand's monthly marketing plan before the money is spent. The AI proposes; a deterministic engine tests every plan against 20 bad months; the owner decides.

## Links
- Live: https://rehearsal-run.fly.dev
- Code: https://github.com/jiwonbabe/rehearsal
- Demo video: [attach rehearsal-demo.mp4]
- X post: [paste the link after posting]

## Description (form field, ~150 words)
A small brand owner plans a month of marketing, spends the budget, and learns in week 3 that it did not work. Rehearsal moves that moment to before the spend.

Muse Spark proposes four directions written for the brand, with parameter ranges, and plays the adversary that invents 20 bad months (creators who never post, a rival's sale, ad costs up 60%, the holiday lift that never comes). It never writes an order count, a revenue figure or a rate. A deterministic Go engine turns the directions into 240 plans, runs each against every bad month, and reports what survived, what each plan dies if, the one assumption that flips the ranking, and where every dollar goes. The owner commits, and Launch creates a paused campaign in a real Meta ad account plus a creator kit and calendar.

Before any plan is shown the engine backtests itself on the brand's own history and shows its hits and misses. Three synthetic demo brands (jewelry, apparel, skincare) run end to end.

## How Muse Spark is used
Five roles through the Meta Model API (text + image input): planner, adversary, check line, profile reader, executor (playbook, creator brief, channel strategist, Korean translation). Every response is validated before it is shown; a sentence quoting a number the engine did not produce is dropped.

## What is real vs synthetic
Real: the Meta Marketing API connection (live reach estimate; paused campaign and ad set created in a real account), public data (Census, Signet, Gupta Media CPM, Pew 2026). Synthetic: the three brands, their histories, mock Instagram insights, fictional creator handles.

## Built
Nov 2–4, 2026, solo, Go stdlib only, no external code. Python reference engine with golden fixtures checks the Go engine on every test run.
