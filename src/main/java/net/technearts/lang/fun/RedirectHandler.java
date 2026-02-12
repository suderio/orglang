package net.technearts.lang.fun;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.*;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

public class RedirectHandler {

    private URL url;

    private RedirectHandler(URL url) {
        this.url = url;
    }

    public static RedirectHandler url(String url) {
        if (!isValidUrl(url))
            throw new RuntimeException(url);
        try {
            return new RedirectHandler(new URI(url).toURL());
        } catch (MalformedURLException | URISyntaxException e) {
            throw new RuntimeException(url);
        }
    }

    // Verifica se uma string é uma URL válida
    private static boolean isValidUrl(String url) {
        return url.matches("^(http|https|ftp|file)://.*$");
    }

    // TODO Operação HTTP
    Object handleHttpOperation(String content) {
        try {
            // Adiciona o conteúdo como um caminho (ou parâmetros GET, dependendo do caso)
            URL fullUrl = new URL(url + content);

            // Abre conexão e lê o recurso
            HttpURLConnection connection = (HttpURLConnection) fullUrl.openConnection();
            connection.setRequestMethod("GET");

            try (BufferedReader reader = new BufferedReader(new InputStreamReader(connection.getInputStream()))) {
                StringBuilder response = new StringBuilder();
                String line;
                while ((line = reader.readLine()) != null) {
                    response.append(line).append("\n");
                }
                return response.toString().strip();
            }
        } catch (IOException e) {
            throw new RuntimeException("Erro ao acessar URL: " + url, e);
        }
    }

    // TODO Operação de arquivo

    /**
     * Create = PUT with a new URI
     *          POST to a base URI returning a newly created URI
     * Read   = GET
     * Update = PUT with an existing URI
     * Delete = DELETE
     * @param content
     * @return
     */
    Object handleFileOperation(String content) {
        Path filePath = getPathFromFileUrl(url.toString());
        try {
            // Escreve no arquivo
            Files.writeString(filePath, content);
            return "File written successfully to " + filePath;
        } catch (IOException e) {
            throw new RuntimeException("Erro ao escrever no arquivo: " + filePath, e);
        }
    }

    public static Path getPathFromFileUrl(String fileUrl) {
        try {
            // Verifica se a URL começa com o esquema "file://"
            if (!fileUrl.startsWith("file://")) {
                throw new IllegalArgumentException("The URL must start with file://");
            }

            // Remove o prefixo "file://"
            String rawPath = fileUrl.substring(7);

            // Identifica o sistema operacional
            String os = System.getProperty("os.name").toLowerCase();

            if (os.contains("win")) {
                // Windows: Remove a barra inicial para caminhos com unidade (ex.: "C:/")
                if (rawPath.matches("^/[a-zA-Z]:.*")) {
                    rawPath = rawPath.substring(1);
                }
                // Converte para um Path no Windows
                return Paths.get(rawPath.replace("/", "\\"));
            } else {
                // Outros sistemas operacionais (Linux, macOS): Diretamente como Unix Path
                return Paths.get(rawPath);
            }
        } catch (Exception e) {
            throw new RuntimeException("Failed to convert file URL to Path: " + fileUrl, e);
        }
    }

    public Object getFile(Path path) throws IOException {
        if (Files.isRegularFile(path)) {
            return Files.readString(path);
        } else if (Files.isDirectory(path)) {
            Table directoryContents = new Table();
            try (DirectoryStream<Path> stream = Files.newDirectoryStream(path)) {
                for (Path entry : stream) {
                    String fileUri = entry.toUri().toString();
                    if (Files.isDirectory(entry)) {
                        directoryContents.put(fileUri, getFile(entry));
                    } else {
                        directoryContents.put(fileUri, null);
                    }
                }
            }
            return directoryContents;
        } else {
            throw new IllegalArgumentException("Path is neither a regular file nor a directory: " + path);
        }
    }
}
