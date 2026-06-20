class CreateScouts < ActiveRecord::Migration[8.1]
  def change
    create_table :scouts do |t|
      t.references :company, null: false, foreign_key: true
      t.string :occupation
      t.string :prefecture
      t.integer :required_experience

      t.timestamps
    end
  end
end
