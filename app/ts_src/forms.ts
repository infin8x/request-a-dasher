$(document).ready(function () {
  $('.phone').mask('(000) 000-0000', { placeholder: "(___) ___-____" });
  $('.money').mask('000.99', { placeholder: "_.__", reverse: true });
  $('.date').each(function(){
    $(this).text(new Date($(this).text()).toLocaleString())
  });
  $('.currentTimezone').text(Intl.DateTimeFormat().resolvedOptions().timeZone);
});

