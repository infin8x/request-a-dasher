"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
jQuery(() => {
    $('.phone').mask('(000) 000-0000', { placeholder: "(___) ___-____" });
    $('.money').mask('000.99', { placeholder: "_.__", reverse: true });
    $('.date').each((i, ele) => {
        $(ele).text(new Date($(ele).text()).toLocaleString());
    });
    $('#externalDeliveryId');
    $('.currentTimezone').text(Intl.DateTimeFormat().resolvedOptions().timeZone);
    $('#moreInfoButton').on('click', () => {
        if ($('#moreInfoButton').text().includes('more')) {
            $('#moreInfoButton').text($('#moreInfoButton').text().replace('more', 'less'));
        }
        else {
            $('#moreInfoButton').text($('#moreInfoButton').text().replace('less', 'more'));
        }
    });
    $('#clearButton').on('click', () => {
        $('#whereFrom').val('');
    });
    // Save from address to localStorage, and get it out again if it's there
    $('#form').on('submit', () => {
        $('#form').find('input[type=submit]').prop('disabled', true);
        localStorage.setItem('whereFrom', $('#whereFrom').val());
    });
});
function initMap() {
    return __awaiter(this, void 0, void 0, function* () {
        let [fromMap, fromMarker] = setUpAutocomplete($('#whereFromMap')[0], $('#whereFrom')[0]);
        let [toMap, toMarker] = setUpAutocomplete($('#whereToMap')[0], $('#whereTo')[0]);
        if ($('#whereTo').val() !== '') { // We're on the deliveries page
            getPlaceAndUpdateMap(fromMap, fromMarker, $('#whereFrom').val());
            getPlaceAndUpdateMap(toMap, toMarker, $('#whereTo').val());
        }
        else { // We're on the request page
            if (localStorage.getItem('whereFrom') !== '') {
                $('#whereFrom').val(localStorage.getItem('whereFrom'));
                getPlaceAndUpdateMap(fromMap, fromMarker, $('#whereFrom').val());
            }
        }
    });
}
function setUpAutocomplete(mapElement, addressElement) {
    const googleMap = new google.maps.Map(mapElement, {
        zoom: 11,
        center: { lat: 47.6073185, lng: -122.3380599 },
        mapTypeControl: false,
        fullscreenControl: false,
        zoomControl: true,
        streetViewControl: false
    });
    const marker = new google.maps.Marker({ map: googleMap, draggable: false });
    const autocomplete = new google.maps.places.Autocomplete(addressElement, {
        fields: ["address_components", "geometry", "name"],
        componentRestrictions: { country: "us" },
    });
    autocomplete.addListener('place_changed', function () {
        marker.setVisible(false);
        const place = autocomplete.getPlace();
        if (!place.geometry) {
            // User entered the name of a Place that was not suggested and
            // pressed the Enter key, or the Place Details request failed.
            window.alert('Couldn\'t find any places that match: \'' + place.name + '\'');
            return;
        }
        renderAddress(googleMap, marker, place);
    });
    return [googleMap, marker];
}
function getPlaceAndUpdateMap(map, marker, address) {
    let service = new google.maps.places.PlacesService(map);
    service.findPlaceFromQuery({ query: address, fields: ["formatted_address", "geometry"] }, (results, status) => {
        if (status === google.maps.places.PlacesServiceStatus.OK && results) {
            renderAddress(map, marker, results[0]);
        }
    });
}
function renderAddress(map, marker, place) {
    map.setCenter(place.geometry.location);
    map.setZoom(17);
    marker.setVisible(true);
    marker.setPosition(place.geometry.location);
}
