class Address
  module Country
    JAPAN = "Japan"
    ALL = [JAPAN].freeze
  end

  attr_reader :country,
              :post_code,
              :prefecture,
              :city,
              :street,
              :building,

  def initialize(postal_code:, prefecture:, city:, street:, building: nil, country:Country::JAPAN)
    if !Country::ALL.include?(country)
      rails ArgumentError, "Invalid country: #{country}"
    end

    @country     = country
    @postal_code = postal_code
    @prefecture  = prefecture
    @city        = city
    @street      = street
    @building    = building
  end

  def full
    case @country
    when Country::JAPAN
      full_japan
    else
      raise NotImplementedError, "Full address for country #{@country} is not implemented"
    end
  end

  private
    def full_japan
      address = "〒#{@postal_code} #{@prefecture} #{@city} #{@street}"
      address += " #{@building}" if @building.present?
      address
    end

end
