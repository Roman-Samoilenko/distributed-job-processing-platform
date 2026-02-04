package com.jobplatform.grpc;

import com.jobplatform.service.JobService;
// ИСПРАВЛЕНИЕ: Импортируем сгенерированные классы напрямую
import com.jobplatform.grpc.UpdateJobStatusRequest;
import com.jobplatform.grpc.UpdateJobStatusResponse;
import com.jobplatform.grpc.JobStatusServiceGrpc;

import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

@Slf4j
@GrpcService
@RequiredArgsConstructor
public class JobStatusGrpcService extends JobStatusServiceGrpc.JobStatusServiceImplBase {
    
    private final JobService jobService;
    
    @Override
    public void updateJobStatus(
            UpdateJobStatusRequest request, // Используем класс напрямую
            StreamObserver<UpdateJobStatusResponse> responseObserver) {
        
        try {
            log.info("Received status update for job {}: {}", request.getJobId(), request.getStatus());
            
            // Сравниваем enum напрямую
            String status = request.getStatus() == UpdateJobStatusRequest.JobStatus.COMPLETED 
                ? "COMPLETED" : "FAILED";
            
            jobService.updateJobStatus(
                request.getJobId(),
                status,
                request.getResult(),
                request.getErrorMessage()
            );
            
            UpdateJobStatusResponse response = UpdateJobStatusResponse.newBuilder()
                    .setSuccess(true)
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (Exception e) {
            log.error("Failed to update job status", e);
            responseObserver.onError(e);
        }
    }
}
