package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"text/template"
)

var (
	addr           = flag.String("web.listen-address", ":9888", "Address on which to expose metrics and web interface.")
	videoBridgeURL = flag.String("videobridge-url", "http://localhost:8888/stats", "Jitsi Videobridge /stats URL to scrape")
	user           = flag.String("web.user", "", "User name for http basic authentication.")
	password       = flag.String("web.password", "", "Password for http basic authentication.")
)

type videoBridgeStats struct {
	Threads                               int     `json:"threads"`
	BitRateDownload                       int     `json:"bit_rate_download"`
	BitRateUpload                         int     `json:"bit_rate_upload"`
	PacketRateDownload                    int     `json:"packet_rate_download"`
	PacketRateUpload                      int     `json:"packet_rate_upload"`
	LossRateDownload                      float64 `json:"loss_rate_download"`
	LossRateUpload                        float64 `json:"loss_rate_upload"`
	JitterAggregate                       float64 `json:"jitter_aggregate"`
	RTTAggregate                          float64 `json:"rtt_aggregate"`
	LargestConference                     int     `json:"largest_conference"`
	ConferenceSizes                       []int   `json:"conference_sizes"`
	AudioChannels                         int     `json:"audiochannels"`
	VideoChannels                         int     `json:"videochannels"`
	Conferences                           int     `json:"conferences"`
	Participants                          int     `json:"participants"`
	Videostreams                          int     `json:"videostreams"`
	TotalLossControlledParticipantSeconds int     `json:"total_loss_controlled_participant_seconds"`
	TotalLossLimitedParticipantSeconds    int     `json:"total_loss_limited_participant_seconds"`
	TotalLossDegradedParticipantSeconds   int     `json:"total_loss_degraded_participant_seconds"`
	TotalConferenceSeconds                int     `json:"total_conference_seconds"`
	TotalConferencesCreated               int     `json:"total_conferences_created"`
	TotalConferencesCompleted             int     `json:"total_conferences_completed"`
	TotalFailedConferences                int     `json:"total_failed_conferences"`
	TotalPartiallyFailedConferences       int     `json:"total_partially_failed_conferences"`
	TotalDataChannelMessagesReceived      int     `json:"total_data_channel_messages_received"`
	TotalDataChannelMessagesSent          int     `json:"total_data_channel_messages_sent"`
	TotalColibriWebSocketMessagesReceived int     `json:"total_colibri_web_socket_messages_received"`
	TotalColibriWebSocketMessagesSent     int     `json:"total_colibri_web_socket_messages_sent"`
	TotalParticipants                     int     `json:"total_participants"`
}

var tpl = template.Must(template.New("stats").Parse(`# HELP jitsi_threads The number of Java threads that the video bridge is using.
# TYPE jitsi_threads gauge
jitsi_threads {{.Threads}}
# HELP jitsi_bit_rate_download The total incoming bitrate for the video bridge in kilobits per second.
# TYPE jitsi_bit_rate_download gauge
jitsi_bit_rate_download {{.BitRateDownload}}
# HELP jitsi_bit_rate_upload The total outgoing bitrate for the video bridge in kilobits per second.
# TYPE jitsi_bit_rate_upload gauge
jitsi_bit_rate_upload {{.BitRateUpload}}
# HELP jitsi_packet_rate_download The total incoming packet rate for the video bridge in packets per second.
# TYPE jitsi_packet_rate_download gauge
jitsi_packet_rate_download {{.PacketRateDownload}}
# HELP jitsi_packet_rate_upload The total outgoing packet rate for the video bridge in packets per second.
# TYPE jitsi_packet_rate_upload gauge
jitsi_packet_rate_upload {{.PacketRateUpload}}
# HELP jitsi_loss_rate_download The fraction of lost incoming RTP packets. This is based on RTP sequence numbers and is relatively accurate.
# TYPE jitsi_loss_rate_download gauge
jitsi_loss_rate_download {{.LossRateDownload}}
# HELP jitsi_loss_rate_upload The fraction of lost outgoing RTP packets. This is based on incoming RTCP Receiver Reports, and an attempt to subtract the fraction of packets that were not sent (i.e. were lost before they reached the bridge). Further, this is averaged over all streams of all users as opposed to all packets, so it is not correctly weighted. This is not accurate, but may be a useful metric nonetheless.
# TYPE jitsi_loss_rate_upload gauge
jitsi_loss_rate_upload {{.LossRateUpload}}
# HELP jitsi_jitter_aggregate Experimental. An average value (in milliseconds) of the jitter calculated for incoming and outgoing streams. This hasn't been tested and it is currently not known whether the values are correct or not.
# TYPE jitsi_jitter_aggregate gauge
jitsi_jitter_aggregate {{.JitterAggregate}}
# HELP jitsi_rtt_aggregate An average value (in milliseconds) of the RTT across all streams.
# TYPE jitsi_rtt_aggregate gauge
jitsi_rtt_aggregate {{.RTTAggregate}}
# HELP jitsi_largest_conference The number of participants in the largest conference currently hosted on the bridge.
# TYPE jitsi_largest_conference gauge
jitsi_largest_conference {{.LargestConference}}
# HELP jitsi_audiochannels The current number of audio channels.
# TYPE jitsi_audiochannels gauge
jitsi_audiochannels {{.AudioChannels}}
# HELP jitsi_videochannels The current number of video channels.
# TYPE jitsi_videochannels gauge
jitsi_videochannels {{.VideoChannels}}
# HELP jitsi_conferences The current number of conferences.
# TYPE jitsi_conferences gauge
jitsi_conferences {{.Conferences}}
# HELP jitsi_participants The current number of participants.
# TYPE jitsi_participants gauge
jitsi_participants {{.Participants}}
# HELP jitsi_videostreams An estimation of the number of current video streams forwarded by the bridge.
# TYPE jitsi_videostreams gauge
jitsi_videostreams {{.Videostreams}}
# HELP jitsi_total_loss_controlled_participant_seconds The total number of participant-seconds that are loss-controlled.
# TYPE jitsi_total_loss_controlled_participant_seconds counter
jitsi_total_loss_controlled_participant_seconds {{.TotalLossControlledParticipantSeconds}}
# HELP jitsi_total_loss_limited_participant_seconds The total number of participant-seconds that are loss-limited.
# TYPE jitsi_total_loss_limited_participant_seconds counter
jitsi_total_loss_limited_participant_seconds {{.TotalLossLimitedParticipantSeconds}}
# HELP jitsi_total_loss_degraded_participant_seconds The total number of participant-seconds that are loss-degraded.
# TYPE jitsi_total_loss_degraded_participant_seconds counter
jitsi_total_loss_degraded_participant_seconds {{.TotalLossDegradedParticipantSeconds}}
# HELP jitsi_total_conference_seconds The sum of the lengths of all completed conferences, in seconds.
# TYPE jitsi_total_conference_seconds counter
jitsi_total_conference_seconds {{.TotalConferenceSeconds}}
# HELP jitsi_total_conferences_created The total number of conferences created on the bridge.
# TYPE jitsi_total_conferences_created counter
jitsi_total_conferences_created {{.TotalConferencesCreated}}
# HELP jitsi_total_conferences_completed The total number of conferences completed on the bridge.
# TYPE jitsi_total_conferences_completed counter
jitsi_total_conferences_completed {{.TotalConferencesCompleted}}
# HELP jitsi_total_failed_conferences The total number of failed conferences on the bridge. A conference is marked as failed when all of its channels have failed. A channel is marked as failed if it had no payload activity.
# TYPE jitsi_total_failed_conferences counter
jitsi_total_failed_conferences {{.TotalFailedConferences}}
# HELP jitsi_total_partially_failed_conferences The total number of partially failed conferences on the bridge. A conference is marked as partially failed when some of its channels has failed. A channel is marked as failed if it had no payload activity.
# TYPE jitsi_total_partially_failed_conferences counter
jitsi_total_partially_failed_conferences {{.TotalPartiallyFailedConferences}}
# HELP jitsi_total_data_channel_messages_received The total number messages received through data channels.
# TYPE jitsi_total_data_channel_messages_received counter
jitsi_total_data_channel_messages_received {{.TotalDataChannelMessagesReceived}}
# HELP jitsi_total_data_channel_messages_sent The total number messages sent through data channels.
# TYPE jitsi_total_data_channel_messages_sent counter
jitsi_total_data_channel_messages_sent {{.TotalDataChannelMessagesSent}}
# HELP jitsi_total_colibri_web_socket_messages_received The total number messages received through COLIBRI web sockets.
# TYPE jitsi_total_colibri_web_socket_messages_received counter
jitsi_total_colibri_web_socket_messages_received {{.TotalColibriWebSocketMessagesReceived}}
# HELP jitsi_total_colibri_web_socket_messages_sent The total number messages sent through COLIBRI web sockets.
# TYPE jitsi_total_colibri_web_socket_messages_sent counter
jitsi_total_colibri_web_socket_messages_sent {{.TotalColibriWebSocketMessagesSent}}
`))

type handler struct {
	sourceURL string
	user      string
	password  string
}

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", h.sourceURL, nil)
	if len(h.user) > 0 && len(h.password) > 0 {
		req.SetBasicAuth(h.user, h.password)
	}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("scrape error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("HTTP Status %d", resp.StatusCode)
		log.Printf("scrape error: %s", msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var stats videoBridgeStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		log.Printf("json decoding error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_ = tpl.Execute(w, &stats)
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	http.Handle("/metrics", handler{sourceURL: *videoBridgeURL, user: *user, password: *password})
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}

	log.Println("Started Jitsi Meet Metrics Exporter")
}
