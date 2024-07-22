package tg

// formatting options:
// https://core.telegram.org/bots/api#formatting-options

const messageHelp = "This bot helps you to find new vacancies on hh.ru based on your search Queries.\n\n" +
	messageAddQuery + "\n\n" + messageRemoveQuery

const messageAddQuery = `To add a new query, send a message to the bot in the following format: 
<code>add: [area: int] [role_id: int] [keywords: string] [experience: (-|0|1-3|3-6|6)]</code>
Example: <code>add: 1 96 golang-разработчик 1-3</code>`

const messageRemoveQuery = `To delete one of the Queries, send the following:
<code>remove: [query_id: int]</code>`

const messageNoQueries = "No active queries found."
