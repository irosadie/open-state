const enumToObject = <T extends { [key: string]: string }>(
  enumObj: T,
): { [key in keyof T]: string } =>
  Object.keys(enumObj).reduce(
    (acc, key) => {
      acc[key as keyof T] = enumObj[key as keyof T]
      return acc
    },
    {} as { [key in keyof T]: string },
  )

export default enumToObject
