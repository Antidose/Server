/**
 * Created by geoff on 7/25/17.
 */
function mapIt (request) {
    let data = JSON.parse(request.responseText);
    let directionsService = new google.maps.DirectionsService;
    let directionsDisplay = new google.maps.DirectionsRenderer({suppressMarkers: true});
    let map = new google.maps.Map(document.getElementById('map'), {});
    let bounds = new google.maps.LatLngBounds();
    let infoWindow = new google.maps.InfoWindow({maxWidth: 450, maxHeight: 500});
    let markers = [];
    let marker, contentString, lat, lng, start, end, overlap, position;

    directionsDisplay.setMap(map);

    //create the markers
    if (data.Incidents.length > 0) {
        for (let incident of data.Incidents) {
            lat = incident.Latitude;
            lng = incident.Longitude;
            position = new google.maps.LatLng(lat, lng);
            //html for the popup info window
            start = "Start Time: " + new Date(incident.Start).toLocaleString();
            end = incident.End.Valid ? "End Time " + new Date(incident.End.String).toLocaleString() : "";
            contentString =
                '<h4>Incident ' + incident.IncID + '</h4>' +
                '<p>' + start + '</p>' +
                '<p>' + end + '</p>';
            marker = new google.maps.Marker({
                map: null,
                position: position,
                title: "Incident",
                groupSize: 1,
                text: contentString,
                animation: google.maps.Animation.BOUNCE
            });
            overlap = false;
            //check if new marker overlaps with another marker
            for (let mark of markers) {
                //add marker info to currenty exsting marker
                if (position.equals(mark.getPosition())) {
                    overlap = true;
                    mark.text = mark.text + '<hr style="border-top: 1px solid #cccccc;" />' + contentString;
                    mark.groupSize = mark.groupSize + 1;
                    if (mark.groupSize > 1) {
                        mark.setLabel(String(mark.groupSize));
                    }
                    break;
                }
            }
            //add marker to map if it does not overlap
            if (!overlap) {
                marker.setMap(map);
                bounds.extend(position); //map will zoom and move so that all markers are on the screen
                map.fitBounds(bounds);
                google.maps.event.addListener(marker, 'mouseover', (function (marker) {
                    return function () {
                        contentString = marker.text;
                        infoWindow.setContent(contentString);
                        infoWindow.open(map, marker);
                        if (getRespondersForIncident(data.Responders, incident) !== 0) {
                            calculateAndDisplayRoute(
                                directionsService,
                                directionsDisplay,
                                marker.position,
                                getRespondersForIncident(data.Responders, incident)
                            );
                        }
                    };
                })(marker));
                markers.push(marker);
            }
        }
    }

    if (data.Responders.length > 0) {
        for (let responder of data.Responders) {
            lat = responder.Latitude;
            lng = responder.Longitude;
            position = new google.maps.LatLng(lat, lng);
            contentString =
                '<h4>Responder ' + responder.Uid + '</h4>' +
                '<p>' + responder.First + ' ' + responder.Last + '</p>' +
                '<p> Responding to: ' + responder.RespondingTo + '</p>';
            marker = new google.maps.Marker({
                map: null,
                position: position,
                title: responder.First,
                groupSize: 1,
                text: contentString,
                icon: 'http://maps.google.com/mapfiles/ms/icons/blue-dot.png'
            });
            overlap = false;
            //check if new marker overlaps with another marker
            for (let mark of markers) {
                //add marker info to currenty exsting marker
                if (position.equals(mark.getPosition())) {
                    overlap = true;
                    mark.text = mark.text + '<hr style="border-top: 1px solid #cccccc;" />' + contentString;
                    mark.groupSize = mark.groupSize + 1;
                    if (mark.groupSize > 1) {
                        mark.setLabel(String(mark.groupSize));
                    }
                    break;
                }
            } //end for
            //add marker to map if it does not overlap
            if (!overlap) {
                marker.setMap(map);
                bounds.extend(position); //map will zoom and move so that all markers are on the screen
                map.fitBounds(bounds);
                google.maps.event.addListener(marker, 'mouseover', (function (marker) {
                    return function () {
                        contentString = marker.text;
                        infoWindow.setContent(contentString);
                        infoWindow.open(map, marker);
                    };
                })(marker));
                markers.push(marker);
            }
        }

        google.maps.event.addListener(map, 'click', (function (marker) {
            return function () {
                infoWindow.close();
            };
        })(marker));

        let options = {
            imagePath: "https://developers.google.com/maps/documentation/javascript/examples/markerclusterer/m",
            maxZoom: 20,
            gridSize: 20
        };

        let markerCluster = new MarkerClusterer(map, markers, options);

        if (bounds.isEmpty()) {
            map.setCenter({lat: 48.463150, lng: -123.312189});
            map.setZoom(5);
        }
    }
}
function calculateAndDisplayRoute(directionsService, directionsDisplay, start, end) {
    directionsService.route({
        origin: start,
        destination: end,
        travelMode: 'WALKING'
    }, function(response, status) {
        if (status === 'OK') {
            directionsDisplay.setDirections(response);
        } else {
            window.alert('Directions request failed due to ' + status);
        }
    });
}

function getRespondersForIncident(responders, incident) {
    for(let responder of responders){
        if (responder.RespondingTo === incident.IncID){
            console.log("directing " + responder.First + " " + responder.Latitude + " " + responder.Longitude);
            console.log("To " + incident.IncID + " " + incident.Latitude + " " + incident.Longitude);
            return new google.maps.LatLng(responder.Latitude, responder.Longitude);
        }
    }
    return 0;
}