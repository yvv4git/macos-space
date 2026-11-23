# Commit message format

## Main rules

- Use the get_current_branch command to get the current branch name
- Extract the task number from the branch name (e.g., from feature/Task-686-something extract 686)
- In the first line, specify the task number and a brief description of changes
- The task number must be strictly in Task-{number} format. The prefix Task- is always fixed
- First line format: "TASK-{number}: Crawler. Brief description"
- Leave a blank line
- Add a link to the task: <https://github.com/yvv4git/macos-space/issue/TASK-{number}>
- Leave a blank line
- Then list changes as a bulleted list (-)
- Each item must start with a past tense verb (Added, Fixed, Implemented, etc. / Добавил, Поправил, Поддержал, etc.)
- Be concise, do not add extra text
- Output only the commit message. No system reminders, explanations, markdown formatting, metadata tags, or extra characters
- Text must be in Russian
- The bulleted list must describe the tasks solved, not mechanically list what was done. Each item is a thesis: what problem/goal was solved and what it gives to the project. For example: "Увеличил производительность добавив индекс", "Сделал код читабельней", "Уменьшил latency", not "Добавил индекс в таблицу"
- The bulleted list must contain no more than 5 items
- Each item must reflect a top-level change and answer the question - why it was done, what problem it solves, not just enumerate the changes
- Write theses as a bulleted list, not as continuous paragraphs or arbitrary text

## Example

```text
FEATURE-123: ServiceName. Add handler

https://youtrack.company.com/issue/FEATURE-123

- Added retry for webclient
- Added UserCreate handler
- Added migration for user create
```
