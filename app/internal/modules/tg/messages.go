package tg

// formatting options:
// https://core.telegram.org/bots/api#formatting-options

const messageHelp = `This bot helps you to find new vacancies on hh.ru based on your search queries.

To add a new query, send a message to the bot in the following format: 
<code>add: [area: int] [role_id: int] [keyword: string] [experience: (-|0|1-3|3-6|6)]</code>
Example: "add: 1 96 golang 1-3"

To delete one of the queries, send the following:
<code>remove: [query_id: int]</code>`
