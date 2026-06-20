class CreateCompanies < ActiveRecord::Migration[8.1]
  def change
    create_table :companies do |t|
      t.string :name
      t.string :industry
      t.string :country
      t.string :postal_code
      t.string :prefecture
      t.string :city
      t.string :street
      t.string :building

      t.timestamps
    end
  end
end
