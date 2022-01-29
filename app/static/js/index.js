$(document).ready(function () {
  $('.phone').mask('(000) 000-0000', { placeholder: "(___) ___-____" });
  $('.money').mask('000.99', { placeholder: "_.__", reverse: true });
});