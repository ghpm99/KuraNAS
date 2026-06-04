SELECT EXISTS (SELECT 1 FROM assistant_conversations WHERE id = $1);
