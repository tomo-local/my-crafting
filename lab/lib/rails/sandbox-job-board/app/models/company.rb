class Company < ApplicationRecord
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

  has_many :jobs
  validates :name, presence: true
end
