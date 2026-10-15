"""Expand one completed synthetic API-created conversation; never targets user databases."""
import json
import sqlite3
import sys

path, conversation_id = sys.argv[1:]
if 'ebk-assistant-e2e-' not in path or not path.endswith('/fixture.db'):
    raise SystemExit('Refusing non-isolated fixture database')
db = sqlite3.connect(path, timeout=30)
db.row_factory = sqlite3.Row
names = [row[0] for row in db.execute("SELECT name FROM sqlite_master WHERE type='table'")]
def table(suffix):
    return '"' + next(name for name in names if name.endswith(suffix)).replace('"', '""') + '"'
messages, conversations, users = table('assistant_message'), table('assistant_conversation'), table('assistant_user_state')
conversation = dict(db.execute(f'SELECT * FROM {conversations} WHERE id=?', (conversation_id,)).fetchone())
rows = [dict(row) for row in db.execute(f'SELECT * FROM {messages} WHERE conversation_id=?', (conversation_id,))]
question = next(row for row in rows if row['role'] == 'user')
answer = next(row for row in rows if row['role'] == 'assistant' and row['run_state'] == 'completed')
base = max(row['id'] for row in rows) + (1 << 40)
old_bytes = sum(row['used_bytes'] for row in rows)
new_bytes = old_bytes
previous = answer['id']
with db:
    for turn in range(2, 121):
        qid, aid = base + turn * 2, base + turn * 2 + 1
        for source, identity, role in [(question, qid, 'user'), (answer, aid, 'assistant')]:
            row = dict(source)
            row.update(id=identity, user_message_id=qid, turn_sequence=turn,
                       client_request_id=f'isolated-history-{turn}-{role}', previous_assistant_message_id=previous,
                       selected_assistant_message_id=aid if role == 'user' else 0)
            row['content'] = f'History fixture turn {turn}:\n' + source['content']
            row['used_bytes'] += len(row['content'].encode()) - len(source['content'].encode())
            new_bytes += row['used_bytes']
            columns = ','.join('"' + key + '"' for key in row)
            db.execute(f'INSERT INTO {messages} ({columns}) VALUES ({",".join("?" for _ in row)})', tuple(row.values()))
        previous = aid
    db.execute(f'UPDATE {conversations} SET turn_count=120,next_turn_sequence=121,context_version=context_version+1,used_bytes=used_bytes+? WHERE id=?', (new_bytes-old_bytes, conversation_id))
    db.execute(f'UPDATE {users} SET used_bytes=used_bytes+? WHERE uid=?', (new_bytes-old_bytes, conversation['uid']))
print(json.dumps({'turns': 120, 'seed': 'completed API-created synthetic question/answer copied only inside isolated SQLite fixture', 'messageBytes': new_bytes}))
