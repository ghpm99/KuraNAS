package com.kuranas.mobile.data.repository;

import com.kuranas.mobile.data.remote.api.FileApi;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.repository.FileRepository;
import com.kuranas.mobile.infra.http.ApiCallback;

public final class FileRepositoryImpl implements FileRepository {

    private final FileApi fileApi;

    public FileRepositoryImpl(FileApi fileApi) {
        this.fileApi = fileApi;
    }

    @Override
    public void getFileTree(int page, int pageSize, ApiCallback<PaginatedResult<FileItem>> callback) {
        fileApi.getTree(page, pageSize, callback);
    }

    @Override
    public void getFilesByPath(String path, int page, int pageSize, ApiCallback<PaginatedResult<FileItem>> callback) {
        fileApi.getByPath(path, page, pageSize, callback);
    }

    @Override
    public void getChildren(int parentId, int page, int pageSize, ApiCallback<PaginatedResult<FileItem>> callback) {
        fileApi.getChildren(parentId, page, pageSize, callback);
    }

    @Override
    public void getImages(int page, int pageSize, String groupBy, ApiCallback<PaginatedResult<FileItem>> callback) {
        fileApi.getImages(page, pageSize, groupBy, callback);
    }

    @Override
    public void getStarredFiles(int page, int pageSize, ApiCallback<PaginatedResult<FileItem>> callback) {
        fileApi.getStarred(page, pageSize, callback);
    }
}
