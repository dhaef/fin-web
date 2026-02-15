const successElements = document.querySelectorAll(".form-success");

successElements.forEach((el) => {
  setTimeout(() => {
    el.remove();
  }, 5000);
});
