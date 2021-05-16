# Inputs〜Coreへのパラメータの受け渡し

```protobuf
syntax = "proto3";
option go_package = "github.com/sacloud/autoscaler/proto";

package autoscaler;

service ScalingService {
    rpc Up(ScalingRequest) returns (ScalingResponse);
    rpc Down(ScalingRequest) returns (ScalingResponse);
    rpc Status(StatusRequest) returns (ScalingResponse);
}

message ScalingRequest {
    string source              = 1;
    string action              = 2;
    string resource_group_name = 3;
}

message StatusRequest {
    string scaling_job_id = 1;
}

message ScalingResponse {
    enum ScalingJobStatus {
        UNKNOWN         = 0;
        ACCEPTED        = 1;
        RUNNING         = 2;
        DONE            = 3;
        CANCELED        = 4;
        FAILED          = 5;
    }
    
    string scaling_job_id   = 1;
    ScalingJobStatus status = 2;
}
```