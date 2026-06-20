class CreateUsers < ActiveRecord::Migration[8.1]
  def change
    create_table :users do |t|
      t.string :name
      t.string :email
      t.string :occupation
      t.integer :desired_salary
      t.string :phone_number
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
