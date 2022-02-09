"use strict";

jQuery(() => {
  $('.phone').mask('(000) 000-0000', { placeholder: "(___) ___-____" });
  $('.money').mask('000.99', { placeholder: "_.__", reverse: true });
  $('.date').each((i, ele) => {
    $(ele).text(new Date($(ele).text()).toLocaleString())
  });
  $('#externalDeliveryId')
  $('.currentTimezone').text(Intl.DateTimeFormat().resolvedOptions().timeZone);

  $('#moreInfoButton').on('click', () => {
    if ($('#moreInfoButton').text().includes('more')) {
      $('#moreInfoButton').text($('#moreInfoButton').text().replace('more', 'less'));
    } else {
      $('#moreInfoButton').text($('#moreInfoButton').text().replace('less', 'more'));
    }
  });

  $('#clearButton').on('click', () => {
    $('#whereFrom').val('');
  });

  // Save from address to localStorage, and get it out again if it's there
  $('#form').on('submit', () => {
    $('#form').find('input[type=submit]').prop('disabled', true);
    localStorage.setItem('whereFrom', <string>$('#whereFrom').val());
  });
});

async function initMap() {
  let [fromMap, fromMarker] = setUpAutocomplete($('#whereFromMap')[0], $('#whereFrom')[0]);
  let [toMap, toMarker] = setUpAutocomplete($('#whereToMap')[0], $('#whereTo')[0]);

  if ($('#whereTo').val() !== '') { // We're on the deliveries page
    getPlaceAndUpdateMap(fromMap, fromMarker, <string>$('#whereFrom').val());
    getPlaceAndUpdateMap(toMap, toMarker, <string>$('#whereTo').val());
  } else { // We're on the request page
    if (localStorage.getItem('whereFrom') !== '') {
      $('#whereFrom').val(localStorage.getItem('whereFrom'));
      getPlaceAndUpdateMap(fromMap, fromMarker, <string>$('#whereFrom').val());
    }
  }
}

function setUpAutocomplete(mapElement: HTMLElement, addressElement: HTMLElement): [google.maps.Map, google.maps.Marker] {
  const googleMap = new google.maps.Map(mapElement, {
    zoom: 11,
    center: { lat: 47.6073185, lng: -122.3380599 }, // 3rd Ave WeWork in Seattle
    mapTypeControl: false,
    fullscreenControl: false,
    zoomControl: true,
    streetViewControl: false
  });
  const marker = new google.maps.Marker({ map: googleMap, draggable: false });
  const autocomplete = new google.maps.places.Autocomplete(<HTMLInputElement>addressElement, {
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

function getPlaceAndUpdateMap(map: google.maps.Map, marker: google.maps.Marker, address: string) {
  let service = new google.maps.places.PlacesService(map);
  service.findPlaceFromQuery({ query: address, fields: ["formatted_address", "geometry"] }, (
    results: google.maps.places.PlaceResult[] | null,
    status: google.maps.places.PlacesServiceStatus
  ) => {
    if (status === google.maps.places.PlacesServiceStatus.OK && results) {
      renderAddress(map, marker, results[0]);
    }
  });
}

function renderAddress(map: google.maps.Map, marker: google.maps.Marker, place: google.maps.places.PlaceResult) {
  map.setCenter(place.geometry.location);
  map.setZoom(17);
  marker.setVisible(true);
  marker.setPosition(place.geometry.location);
}