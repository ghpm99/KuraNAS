-- Tiering quente/frio (task 13): separa o path lógico (identidade do arquivo,
-- o que a navegação mostra) da localização física dos bytes. NULL significa
-- "os bytes estão no próprio path" — o caso de todos os arquivos não migrados.
ALTER TABLE home_file
ADD COLUMN physical_path TEXT NULL;
