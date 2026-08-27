const objectToForm = <T extends Record<string, unknown>>(
  obj: T,
  form = new FormData(),
  namespace = "",
) => {
  if (!obj) {
    return form
  }
  for (const key in obj) {
    const value = obj[key]

    if (
      !Object.prototype.hasOwnProperty.call(obj, key) ||
      value === undefined
    ) {
      continue
    }
    const formKey = namespace ? `${namespace}[${key}]` : key

    if (value instanceof File || value instanceof Blob) {
      form.append(formKey, value as Blob)
    } else if (typeof value === "boolean") {
      form.append(formKey, String(value ? 1 : 0))
    } else if (typeof value === "object" && !(value instanceof File)) {
      objectToForm(value as Record<string, unknown>, form, formKey)
    } else {
      form.append(formKey, String(value))
    }
  }
  return form
}

export default objectToForm
