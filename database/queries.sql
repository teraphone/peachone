SELECT users.id, users.name, group_users.created_at, group_users.group_role_id
FROM group_users
JOIN users
ON group_users.user_id = users.id
WHERE group_id = 1;