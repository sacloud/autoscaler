var SEVERITY_COLORS = [
    '#97AAB3', // Not classified.
    '#7499FF', // Information.
    '#FFC859', // Warning.
    '#FFA059', // Average.
    '#E97659', // High.
    '#E45959', // Disaster.
    '#009900'  // Resolved.
];

function stringTruncate(str, len) {
    return str.length > len ? str.substring(0, len - 3) + '...' : str;
}

try {
    Zabbix.Log(4, '[ sacloud/AutoScaler Webhook ] Executed with params: ' + value);

    var params = JSON.parse(value);

    if (!params.autoscaler_endpoint) {
        throw 'Cannot get autoscaler_endpoint';
    }

    if (['up', 'down'].indexOf(params.autoscaler_event_type) === -1) {
        throw 'Incorrect "autoscaler_event_type" parameter given: "' + params.autoscaler_event_type + '".\nMust be "up" or "down".';
    }

    if (!params.autoscaler_source) {
        params.autoscaler_source = "default";
    }
    if (!params.autoscaler_resource_name) {
        params.autoscaler_resource_name = "default";
    }
    params.autoscaler_endpoint = (params.autoscaler_endpoint.endsWith('/'))
        ? params.autoscaler_endpoint.slice(0, -1) : params.autoscaler_endpoint;
    params.autoscaler_endpoint =
        params.autoscaler_endpoint + "/" + params.autoscaler_event_type +
        "?source=" + params.autoscaler_source +
        "&resource_name=" + params.autoscaler_resource_name;

    if (!!params.autoscaler_desired_state_name) {
        params.autoscaler_endpoint = params.autoscaler_endpoint + "&desired-state-name=" + params.autoscaler_desired_state_name;
    }

    params.zabbix_url = (params.zabbix_url.endsWith('/'))
        ? params.zabbix_url.slice(0, -1) : params.zabbix_url;

    if ([0, 1, 2, 3].indexOf(parseInt(params.event_source)) === -1) {
        throw 'Incorrect "event_source" parameter given: "' + params.event_source + '".\nMust be 0-3.';
    }

    // Set params to true for non trigger-based events.
    if (params.event_source !== '0') {
        params.use_default_message = 'true';
        params.event_nseverity = '0';
    }

    // Check {EVENT.VALUE} for trigger-based and internal events.
    if (params.event_value !== '0' && params.event_value !== '1'
        && (params.event_source === '0' || params.event_source === '3')) {
        throw 'Incorrect "event_value" parameter given: "' + params.event_value + '".\nMust be 0 or 1.';
    }

    // Check {EVENT.UPDATE.STATUS} only for trigger-based events.
    if (params.event_update_status !== '0' && params.event_update_status !== '1' && params.event_source === '0') {
        throw 'Incorrect "event_update_status" parameter given: "' + params.event_update_status + '".\nMust be 0 or 1.';
    }

    if (params.event_value == 0) {
        params.event_nseverity = '6';
    }

    if (!SEVERITY_COLORS[params.event_nseverity]) {
        throw 'Incorrect "event_nseverity" parameter given: ' + params.event_nseverity + '\nMust be 0-5.';
    }

    var color = parseInt(SEVERITY_COLORS[params.event_nseverity].replace('#', ''), 16),
        fields = [],
        body = {
            embeds: [
                {
                    color: color || 0,
                    url: (params.event_source === '0')
                        ? params.zabbix_url + '/tr_events.php?triggerid=' + params.trigger_id +
                        '&eventid=' + params.event_id
                        : params.zabbix_url
                }
            ]
        };

    // Default message from {ALERT.MESSAGE}.
    if (params.use_default_message.toLowerCase() == 'true') {
        body.embeds[0].title = stringTruncate(params.alert_subject, 256);
        body.embeds[0].description = stringTruncate(params.alert_message, 2048);
    }
    else {
        fields.push(
            {
                name: 'Host',
                value: params.host_name + ' [' + params.host_ip + ']'
            }
        );

        // Resolved message.
        if (params.event_value == 0 && params.event_update_status == 0) {
            body.embeds[0].title = stringTruncate('OK: ' + params.event_name, 256);
            fields.push(
                {
                    name: 'Recovery time',
                    value: params.event_recovery_time + ' ' + params.event_recovery_date,
                    inline: 'True'
                }
            );
        }

        // Problem message.
        else if (params.event_value == 1 && params.event_update_status == 0) {
            body.embeds[0].title = stringTruncate('PROBLEM: ' + params.event_name, 256);
            fields.push(
                {
                    name: 'Event time',
                    value: params.event_time + ' ' + params.event_date,
                    inline: 'True'
                }
            );
        }

        // Update message.
        else if (params.event_update_status == 1) {
            body.embeds[0].title = stringTruncate('UPDATE: ' + params.event_name, 256);
            body.embeds[0].description = params.event_update_user + ' ' + params.event_update_action + '.';

            if (params.event_update_message) {
                body.embeds[0].description += ' Comment:\n>>> ' + params.event_update_message;
            }

            body.embeds[0].description = stringTruncate(body.embeds[0].description, 2048);

            fields.push(
                {
                    name: 'Event update time',
                    value: params.event_update_time + ' ' + params.event_update_date,
                    inline: 'True'
                }
            );
        }

        fields.push(
            {
                name: 'Severity',
                value: params.event_severity,
                inline: 'True'
            }
        );

        if (params.event_opdata) {
            fields.push(
                {
                    name: 'Operational data',
                    value: stringTruncate(params.event_opdata, 1024),
                    inline: 'True'
                }
            );
        }

        if (params.event_value == 1 && params.event_update_status == 0 && params.trigger_description) {
            fields.push(
                {
                    name: 'Trigger description',
                    value: stringTruncate(params.trigger_description, 1024)
                }
            );
        }

        body.embeds[0].footer = {
            text: 'Event ID: ' + params.event_id
        };

        if (params.event_tags) {
            body.embeds[0].footer.text += '\nEvent tags: ' + params.event_tags;
        }
        body.embeds[0].footer.text = stringTruncate(body.embeds[0].footer.text, 2048);
    }

    if (fields.length > 0) {
        body.embeds[0].fields = fields;
    }

    var req = new CurlHttpRequest();

    if (typeof params.HTTPProxy === 'string' && params.HTTPProxy.trim() !== '') {
        req.SetProxy(params.HTTPProxy);
    }

    req.AddHeader('Content-Type: application/json');

    var resp = req.Post(params.autoscaler_endpoint, JSON.stringify(body)),
        data = JSON.parse(resp);

    Zabbix.Log(4, '[ sacloud/AutoScaler Webhook ] JSON: ' + JSON.stringify(body));
    Zabbix.Log(4, '[ sacloud/AutoScaler Webhook ] Response: ' + resp);

    if (data.id) {
        return resp;
    }
    else {
        var message = ((typeof data.message === 'string') ? data.message : 'Unknown error');

        Zabbix.Log(3, '[ sacloud/AutoScaler Webhook ] FAILED with response: ' + resp);
        throw message + '. For more details check zabbix server log.';
    }
}
catch (error) {
    Zabbix.Log(3, '[ sacloud/AutoScaler Webhook ] ERROR: ' + error);
    throw 'Sending failed: ' + error;
}