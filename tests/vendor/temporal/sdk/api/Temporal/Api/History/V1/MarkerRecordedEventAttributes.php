<?php
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: temporal/api/history/v1/message.proto

namespace Temporal\Api\History\V1;

use Google\Protobuf\Internal\GPBType;
use Google\Protobuf\Internal\RepeatedField;
use Google\Protobuf\Internal\GPBUtil;

/**
 * Generated from protobuf message <code>temporal.api.history.v1.MarkerRecordedEventAttributes</code>
 */
class MarkerRecordedEventAttributes extends \Google\Protobuf\Internal\Message
{
    /**
     * Generated from protobuf field <code>string marker_name = 1;</code>
     */
    protected $marker_name = '';
    /**
     * Generated from protobuf field <code>map<string, .temporal.api.common.v1.Payloads> details = 2;</code>
     */
    private $details;
    /**
     * Generated from protobuf field <code>int64 workflow_task_completed_event_id = 3;</code>
     */
    protected $workflow_task_completed_event_id = 0;
    /**
     * Generated from protobuf field <code>.temporal.api.common.v1.Header header = 4;</code>
     */
    protected $header = null;
    /**
     * Generated from protobuf field <code>.temporal.api.failure.v1.Failure failure = 5;</code>
     */
    protected $failure = null;

    /**
     * Constructor.
     *
     * @param array $data {
     *     Optional. Data for populating the Message object.
     *
     *     @type string $marker_name
     *     @type array|\Google\Protobuf\Internal\MapField $details
     *     @type int|string $workflow_task_completed_event_id
     *     @type \Temporal\Api\Common\V1\Header $header
     *     @type \Temporal\Api\Failure\V1\Failure $failure
     * }
     */
    public function __construct($data = NULL) {
        \GPBMetadata\Temporal\Api\History\V1\Message::initOnce();
        parent::__construct($data);
    }

    /**
     * Generated from protobuf field <code>string marker_name = 1;</code>
     * @return string
     */
    public function getMarkerName()
    {
        return $this->marker_name;
    }

    /**
     * Generated from protobuf field <code>string marker_name = 1;</code>
     * @param string $var
     * @return $this
     */
    public function setMarkerName($var)
    {
        GPBUtil::checkString($var, True);
        $this->marker_name = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>map<string, .temporal.api.common.v1.Payloads> details = 2;</code>
     * @return \Google\Protobuf\Internal\MapField
     */
    public function getDetails()
    {
        return $this->details;
    }

    /**
     * Generated from protobuf field <code>map<string, .temporal.api.common.v1.Payloads> details = 2;</code>
     * @param array|\Google\Protobuf\Internal\MapField $var
     * @return $this
     */
    public function setDetails($var)
    {
        $arr = GPBUtil::checkMapField($var, \Google\Protobuf\Internal\GPBType::STRING, \Google\Protobuf\Internal\GPBType::MESSAGE, \Temporal\Api\Common\V1\Payloads::class);
        $this->details = $arr;

        return $this;
    }

    /**
     * Generated from protobuf field <code>int64 workflow_task_completed_event_id = 3;</code>
     * @return int|string
     */
    public function getWorkflowTaskCompletedEventId()
    {
        return $this->workflow_task_completed_event_id;
    }

    /**
     * Generated from protobuf field <code>int64 workflow_task_completed_event_id = 3;</code>
     * @param int|string $var
     * @return $this
     */
    public function setWorkflowTaskCompletedEventId($var)
    {
        GPBUtil::checkInt64($var);
        $this->workflow_task_completed_event_id = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>.temporal.api.common.v1.Header header = 4;</code>
     * @return \Temporal\Api\Common\V1\Header
     */
    public function getHeader()
    {
        return isset($this->header) ? $this->header : null;
    }

    public function hasHeader()
    {
        return isset($this->header);
    }

    public function clearHeader()
    {
        unset($this->header);
    }

    /**
     * Generated from protobuf field <code>.temporal.api.common.v1.Header header = 4;</code>
     * @param \Temporal\Api\Common\V1\Header $var
     * @return $this
     */
    public function setHeader($var)
    {
        GPBUtil::checkMessage($var, \Temporal\Api\Common\V1\Header::class);
        $this->header = $var;

        return $this;
    }

    /**
     * Generated from protobuf field <code>.temporal.api.failure.v1.Failure failure = 5;</code>
     * @return \Temporal\Api\Failure\V1\Failure
     */
    public function getFailure()
    {
        return isset($this->failure) ? $this->failure : null;
    }

    public function hasFailure()
    {
        return isset($this->failure);
    }

    public function clearFailure()
    {
        unset($this->failure);
    }

    /**
     * Generated from protobuf field <code>.temporal.api.failure.v1.Failure failure = 5;</code>
     * @param \Temporal\Api\Failure\V1\Failure $var
     * @return $this
     */
    public function setFailure($var)
    {
        GPBUtil::checkMessage($var, \Temporal\Api\Failure\V1\Failure::class);
        $this->failure = $var;

        return $this;
    }

}
