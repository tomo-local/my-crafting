class User < ApplicationRecord
  composed_of :address,
    class_name: "Address",
    mapping: [
      postal_code: :postal_code,
      prefecture: :prefecture,
      city: :city,
      street: :street,
      building: :building,
      country: :country
    ]

  attr_accessor :name, :email

  has_many :entries
end
