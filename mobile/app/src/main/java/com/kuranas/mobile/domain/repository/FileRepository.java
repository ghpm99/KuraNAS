package com.kuranas.mobile.domain.repository;

import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.infra.http.ApiCallback;

public interface FileRepository {

    void getFileTree(int page, int pageSize, ApiCallback<PaginatedResult<FileItem>> callback);

    void getFilesByPath(String path, int page, int pageSize, ApiCallback<PaginatedResult<FileItem>> callback);

    void getChildren(int parentId, int page, int pageSize, ApiCallback<PaginatedResult<FileItem>> callback);

    void getImages(int page, int pageSize, String groupBy, ApiCallback<PaginatedResult<FileItem>> callback);

    void getStarredFiles(int page, int pageSize, ApiCallback<PaginatedResult<FileItem>> callback);
}
